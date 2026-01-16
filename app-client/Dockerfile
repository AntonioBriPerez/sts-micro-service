FROM python:3.9-slim

WORKDIR /app

# Instalamos dependencias
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copiamos el código
COPY main.py .

# Exponemos el puerto 3000
EXPOSE 3000

# Arrancamos
CMD ["python", "main.py"]
