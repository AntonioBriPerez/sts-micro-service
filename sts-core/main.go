package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- ESTRUCTURAS DE DATOS ---
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	signKey      *rsa.PrivateKey
	publicKeyPEM string
	usersDb      = map[string]User{}
	mutex        sync.Mutex
)

// --- FUNCIÓN DE AYUDA PARA DEPURAR RUTAS ---
func debugPath(path string) {
	fmt.Printf("DEBUG: Inspeccionando ruta '%s'...\n", path)
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("   -> ERROR: No se puede acceder a %s: %v\n", path, err)
		return
	}
	if fileInfo.IsDir() {
		fmt.Printf("   -> Es un DIRECTORIO. Contenido:\n")
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			fmt.Printf("      - %s\n", e.Name())
		}
	} else {
		fmt.Printf("   -> Es un ARCHIVO. Tamaño: %d bytes\n", fileInfo.Size())
	}
}

// --- MAIN ---
func main() {
	fmt.Println("--- INICIANDO STS (MODO DEBUG KUBERNETES) ---")

	// 1. DIAGNÓSTICO DE ENTORNO
	cwd, _ := os.Getwd()
	fmt.Printf("DEBUG: Directorio de trabajo actual (PWD): %s\n", cwd)
	
	// Intentamos listar la raíz y la carpeta keys para ver si el volumen se montó
	debugPath("/")
	debugPath("/keys") 

	// 2. INTENTO DE CARGA DE CLAVE (Estrategia a prueba de balas)
	// En K8s montamos el volumen en "/keys", así que la ruta absoluta es "/keys/sts_privada.pem"
	// En local a veces usamos "../keys". Probaremos ambas.
	possiblePaths := []string{
		"/keys/sts_privada.pem",      // Ruta estándar K8s
		"../keys/sts_privada.pem",    // Ruta desarrollo local
		"./keys/sts_privada.pem",     // Ruta relativa
	}

	var keyBytes []byte
	var err error
	var loadedPath string

	for _, p := range possiblePaths {
		fmt.Printf("DEBUG: Intentando leer clave en: %s\n", p)
		keyBytes, err = os.ReadFile(p)
		if err == nil {
			loadedPath = p
			break // ¡La encontramos!
		}
	}

	if len(keyBytes) == 0 {
		fmt.Println("FATAL: No se pudo encontrar 'sts_privada.pem' en ninguna ruta esperada.")
		// Hacemos un sleep antes de morir para que te de tiempo a ver los logs si K8s reinicia muy rápido
		time.Sleep(10 * time.Second) 
		os.Exit(1)
	}

	fmt.Printf("INFO: Clave cargada exitosamente desde: %s\n", loadedPath)

	// 3. PARSEO DE CLAVE
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		fmt.Printf("FATAL: El archivo existe pero no es una clave RSA válida: %v\n", err)
		os.Exit(1)
	}

	// Derivar pública
	pubASN1, _ := x509.MarshalPKIXPublicKey(&signKey.PublicKey)
	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1}))

	// 4. LEVANTAR SERVIDOR
	http.HandleFunc("/public-key", publicKeyHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)

	fmt.Println("INFO: Servidor listo y escuchando en :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("FATAL: Error al levantar servidor HTTP: %v\n", err)
		os.Exit(1)
	}
}

// --- HANDLERS (Iguales que antes) ---
func publicKeyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"public_key": publicKeyPEM})
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Usa POST", http.StatusMethodNotAllowed)
		return
	}
	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	mutex.Lock()
	usersDb[newUser.Username] = newUser
	mutex.Unlock()
	fmt.Printf("DB: Usuario registrado -> %s\n", newUser.Username)
	w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Usa POST", http.StatusMethodNotAllowed)
		return
	}
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	mutex.Lock()
	user, exists := usersDb[creds.Username]
	mutex.Unlock()

	if !exists || user.Password != creds.Password {
		http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
		return
	}
	claims := jwt.MapClaims{
		"sub":  user.Username,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 1).Unix(),
		"iss":  "sts-service-debug",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString(signKey)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	fmt.Printf("ACCESS: Login exitoso para %s\n", user.Username)
}
