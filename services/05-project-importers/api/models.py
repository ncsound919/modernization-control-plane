from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import List, Optional

from pydantic import BaseModel, Field


class FormatStatus(str, Enum):
    supported = "supported"
    experimental = "experimental"
    deprecated = "deprecated"


class JobStatus(str, Enum):
    pending = "pending"
    processing = "processing"
    completed = "completed"
    failed = "failed"


class ImportFormat(BaseModel):
    name: str
    extension: str
    description: str
    target_format: str
    status: FormatStatus


class ImportJob(BaseModel):
    id: str
    format: str
    status: JobStatus
    created_at: datetime
    completed_at: Optional[datetime] = None
    input_filename: str
    output_url: Optional[str] = None
    output_data: Optional[str] = None
    warnings: List[str] = Field(default_factory=list)
    errors: List[str] = Field(default_factory=list)


class FLAImportRequest(BaseModel):
    filename: str
    content_base64: str
    target_format: str = Field(default="lottie", pattern="^(lottie|svg|canvas)$")


class EAGLEImportRequest(BaseModel):
    filename: str
    content_base64: str
    file_type: str = Field(default="brd", pattern="^(brd|sch)$")
    target_format: str = Field(default="kicad", pattern="^kicad$")
