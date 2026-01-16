from fastapi import FastAPI, Header, HTTPException
import jwt
import requests
import uvicorn
import os

app = FastAPI()

# CONFIGURACIÓN
# En una red Docker, llamaremos al otro contenedor por su nombre: "sts-instance"
STS_URL = os.getenv('STS_URL', "http://sts-service:80/public-key")

def get_sts_public_key():
    """
    Llama al STS para obtener su clave pública actualizada.
    En producción, esto debería cachearse para no llamar en cada petición.
    """
    try:
        response = requests.get(STS_URL)
        if response.status_code == 200:
            return response.json().get("public_key")
    except Exception as e:
        print(f"Error conectando al STS: {e}")
    return None

@app.get("/")
def home():
    return {"message": "Soy la App Tonta. Usa /secreto con un token."}

@app.get("/secreto")
def protected_route(authorization: str = Header(None)):
    # 1. Verificar que hay cabecera
    if not authorization:
        raise HTTPException(status_code=401, detail="Falta el header Authorization")

    # 2. Limpiar el token (Quitar 'Bearer ')
    try:
        scheme, token = authorization.split()
        if scheme.lower() != 'bearer':
            raise HTTPException(status_code=401, detail="Formato inválido. Usa 'Bearer <token>'")
    except ValueError:
        raise HTTPException(status_code=401, detail="Header mal formado")

    # 3. Obtener la clave pública del STS (Confianza)
    public_key_pem = get_sts_public_key()
    if not public_key_pem:
        raise HTTPException(status_code=503, detail="El STS no responde, no puedo validarte.")

    # 4. Validar la Firma (El momento de la verdad)
    try:
        # PyJWT verifica la firma, la expiración y el formato
        payload = jwt.decode(token, public_key_pem, algorithms=["RS256"])
        
        # Si llegamos aquí, el token es AUTÉNTICO.
        return {
            "status": "ACCESO CONCEDIDO",
            "data_secreta": "El código nuclear es 1234",
            "usuario_validado": payload.get("sub"),
            "rol_detectado": payload.get("role")
        }

    except jwt.ExpiredSignatureError:
        raise HTTPException(status_code=401, detail="El token ha caducado")
    except jwt.InvalidTokenError as e:
        raise HTTPException(status_code=401, detail=f"Token inválido: {str(e)}")

if __name__ == "__main__":
    # Escuchamos en el puerto 3000 para no chocar con el 8080 del STS
    uvicorn.run(app, host="0.0.0.0", port=3000)
