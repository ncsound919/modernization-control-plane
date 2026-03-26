import os

import uvicorn

from orchestrator.app import app  # noqa: F401  (re-export for uvicorn)

if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8080))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
