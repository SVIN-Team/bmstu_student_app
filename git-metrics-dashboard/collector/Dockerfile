FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY status_analyzer.py .
COPY collect_metrics.py .

CMD ["python", "collect_metrics.py"]
