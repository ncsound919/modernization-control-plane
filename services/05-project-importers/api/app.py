import base64
import json
import uuid
from datetime import datetime, timezone
from typing import Dict, List

from fastapi import FastAPI, HTTPException

from api.models import (
    EAGLEImportRequest,
    FLAImportRequest,
    FormatStatus,
    ImportFormat,
    ImportJob,
    JobStatus,
)
from importers.eagle import EAGLEParser, KiCadConverter
from importers.fla import FLAParser, LottieConverter

app = FastAPI(
    title="Project Importers Service",
    description="Parsers and converters for discontinued file formats (.FLA, .BRD, .SCH)",
    version="1.0.0",
)

_jobs: Dict[str, ImportJob] = {}

SUPPORTED_FORMATS: List[ImportFormat] = [
    ImportFormat(
        name="Adobe Animate FLA",
        extension=".fla",
        description="Adobe Animate project file (discontinued March 2026)",
        target_format="lottie|svg|canvas",
        status=FormatStatus.supported,
    ),
    ImportFormat(
        name="Autodesk EAGLE Board",
        extension=".brd",
        description="EAGLE PCB layout file (discontinued June 2026)",
        target_format="kicad",
        status=FormatStatus.supported,
    ),
    ImportFormat(
        name="Autodesk EAGLE Schematic",
        extension=".sch",
        description="EAGLE schematic file (discontinued June 2026)",
        target_format="kicad",
        status=FormatStatus.supported,
    ),
]


@app.get("/health")
def health_check():
    return {"status": "healthy", "service": "project-importers", "version": "1.0.0"}


@app.get("/api/v1/formats", response_model=List[ImportFormat])
def list_formats():
    return SUPPORTED_FORMATS


@app.post("/api/v1/import/fla", response_model=ImportJob, status_code=202)
def import_fla(request: FLAImportRequest):
    job_id = str(uuid.uuid4())
    now = datetime.now(timezone.utc)
    job = ImportJob(
        id=job_id,
        format="fla",
        status=JobStatus.processing,
        created_at=now,
        input_filename=request.filename,
    )
    _jobs[job_id] = job

    try:
        raw = base64.b64decode(request.content_base64)
        parser = FLAParser()
        fla_data = parser.parse(raw)
        converter = LottieConverter()
        output = converter.convert(fla_data, target_format=request.target_format)

        job.status = JobStatus.completed
        job.completed_at = datetime.now(timezone.utc)
        job.output_data = json.dumps(output) if isinstance(output, dict) else output
        job.output_url = f"/api/v1/jobs/{job_id}/output"
        job.warnings = fla_data.get("warnings", [])
    except Exception as exc:
        job.status = JobStatus.failed
        job.completed_at = datetime.now(timezone.utc)
        job.errors = [str(exc)]

    return job


@app.get("/api/v1/import/fla/{job_id}", response_model=ImportJob)
def get_fla_job(job_id: str):
    job = _jobs.get(job_id)
    if job is None or job.format != "fla":
        raise HTTPException(status_code=404, detail="FLA import job not found")
    return job


@app.post("/api/v1/import/eagle", response_model=ImportJob, status_code=202)
def import_eagle(request: EAGLEImportRequest):
    job_id = str(uuid.uuid4())
    now = datetime.now(timezone.utc)
    job = ImportJob(
        id=job_id,
        format=request.file_type,
        status=JobStatus.processing,
        created_at=now,
        input_filename=request.filename,
    )
    _jobs[job_id] = job

    try:
        raw = base64.b64decode(request.content_base64)
        parser = EAGLEParser()
        eagle_data = parser.parse(raw, file_type=request.file_type)
        converter = KiCadConverter()
        output = converter.convert(eagle_data, file_type=request.file_type)

        job.status = JobStatus.completed
        job.completed_at = datetime.now(timezone.utc)
        job.output_data = output
        job.output_url = f"/api/v1/jobs/{job_id}/output"
        job.warnings = eagle_data.get("warnings", [])
    except Exception as exc:
        job.status = JobStatus.failed
        job.completed_at = datetime.now(timezone.utc)
        job.errors = [str(exc)]

    return job


@app.get("/api/v1/import/eagle/{job_id}", response_model=ImportJob)
def get_eagle_job(job_id: str):
    job = _jobs.get(job_id)
    if job is None or job.format not in ("brd", "sch"):
        raise HTTPException(status_code=404, detail="EAGLE import job not found")
    return job


@app.get("/api/v1/jobs", response_model=List[ImportJob])
def list_jobs():
    return list(_jobs.values())


@app.get("/api/v1/jobs/{job_id}", response_model=ImportJob)
def get_job(job_id: str):
    job = _jobs.get(job_id)
    if job is None:
        raise HTTPException(status_code=404, detail="Job not found")
    return job
