# Stage 1: Builder (install dependencies)
FROM alpine:latest

RUN apk add --no-cache python3 py3-pip

WORKDIR /app
#RUN apk add --no-cache python3 py3-pip g++ musl-dev python3-dev

COPY requirements.txt .
COPY *.py .
COPY assets ./assets
RUN pip install --no-cache-dir -r requirements.txt --break-system-packages
RUN apk del py3-pip


ENTRYPOINT ["streamlit", "run", "main.py", "--server.port=8501", "--server.address=0.0.0.0"]