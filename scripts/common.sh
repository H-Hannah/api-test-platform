# Shared helpers for shell scripts (source after loading .env).

# BASE_URL: explicit env wins; else derive from ADDR (e.g. :8081 -> http://localhost:8081).
resolve_base_url() {
  if [ -n "${BASE_URL:-}" ]; then
    return
  fi
  if [ -n "${ADDR:-}" ]; then
    case "$ADDR" in
      :*) BASE_URL="http://localhost${ADDR}" ;;
      *) BASE_URL="http://${ADDR}" ;;
    esac
  else
    BASE_URL="http://localhost:8080"
  fi
}
