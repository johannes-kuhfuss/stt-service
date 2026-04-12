
# Speech-To-Text Service

The Speech-To-Text (STT) Service wraps Speaches.ai to create speech from an existing audio file.

## Configuration (Environment Variables)

- STT_PATH: Folder where audio files are stored, e.g. "/upload" (needs to mounted as a volume in the container)
- SPEACHES_HOST: host name or IP where to find Speaches, e.g. "speaches"
- OTEL_EXPORTER_OTLP_ENDPOINT: endpoint to send OTEL data to, e.g. "<http://192.168.1.100:4317/>"
