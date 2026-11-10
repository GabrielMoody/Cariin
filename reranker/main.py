from dotenv import load_dotenv
import os
from typing import Any, Optional
import sys

from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel, Field

from sentence_transformers import CrossEncoder
from huggingface_hub import snapshot_download
import torch
import logging
from tqdm.auto import tqdm

torch.set_num_threads(4)
torch.set_num_interop_threads(2)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)s | %(name)s | %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
    force=True,
)

load_dotenv()

MODEL_NAME = "jinaai/jina-reranker-v2-base-multilingual"

MODEL_REVISION = "9cfeff2df7d40d1b78e75e5e9cebec92a99813c9"

MODEL_CACHE_DIR = "/app/.cache/huggingface"

HF_TOKEN = os.getenv("HF_TOKEN")

MAX_DOCUMENTS = 5000
MAX_QUERY_LENGTH = 1000
MAX_DOCUMENT_TEXT_LENGTH = 8000


@asynccontextmanager
async def lifespan(app: FastAPI):
    model_path = snapshot_download(
        repo_id=MODEL_NAME,
        revision=MODEL_REVISION,
        cache_dir=MODEL_CACHE_DIR,
        token=HF_TOKEN,
    )

    app.state.model = CrossEncoder(
        model_path,
        max_length=512,
        device="cpu",
        trust_remote_code=True,
    )
    app.state.model.model.float()

    yield

    app.state.model = None


app = FastAPI(
    title="Cariin Reranker API",
    version="1.0.0",
    lifespan=lifespan,
)


class RerankDocument(BaseModel):
    id: int
    title: str = ""
    url: str = ""
    text: str = ""


class RerankRequest(BaseModel):
    query: str = Field(min_length=1, max_length=MAX_QUERY_LENGTH)
    documents: list[RerankDocument] = Field(
        min_length=1,
        max_length=MAX_DOCUMENTS,
    )


class RankedDocument(BaseModel):
    id: int
    score: float
    rank: int


class RerankResponse(BaseModel):
    query: str
    model: str
    results: list[RankedDocument]


@app.get("/")
def root() -> dict[str, str]:
    return {
        "service": "Cariin Reranker API",
        "model": MODEL_NAME,
    }


@app.get("/health")
def health(request: Request) -> dict[str, Any]:
    model = getattr(request.app.state, "model", None)

    if model is None:
        raise HTTPException(
            status_code=503,
            detail="Reranker model is not ready",
        )

    return {
        "status": "ok",
        "model": MODEL_NAME,
    }

logger = logging.getLogger("reranker")

@app.post("/rerank", response_model=RerankResponse)
def rerank(
    payload: RerankRequest,
    request: Request,
) -> RerankResponse:
    model = getattr(request.app.state, "model", None)

    if model is None:
        raise HTTPException(
            status_code=503,
            detail="Reranker model is not ready",
        )

    document_ids = [doc.id for doc in payload.documents]

    if len(document_ids) != len(set(document_ids)):
        raise HTTPException(
            status_code=400,
            detail="Document IDs must be unique",
        )

    total_documents = len(payload.documents)

    logger.info(
        "Starting reranking: query=%r, documents=%d",
        payload.query,
        total_documents,
    )

    passages = [
        (
            f"Title: {doc.title}\n"
            f"URL: {doc.url}\n"
            f"Content: {doc.text}"
        )[:MAX_DOCUMENT_TEXT_LENGTH]
        for doc in payload.documents
    ]

    pairs = [
        (payload.query, passage)
        for passage in passages
    ]

    try:
        batch_size = 2
        scores = []

        total_batches = (
            total_documents + batch_size - 1
        ) // batch_size

        for start in tqdm(
            range(0, total_documents, batch_size),
            total=total_batches,
            desc="Reranking documents",
            unit="batch",
            dynamic_ncols=True,
        ):
            batch = pairs[start:start + batch_size]

            batch_scores = model.predict(
                batch,
                batch_size=batch_size,
                show_progress_bar=False,
            )

            scores.extend(batch_scores)

        logger.info(
            "Reranking completed: %d/%d documents processed",
            len(scores),
            total_documents,
        )

        scored_documents = [
            {
                "id": doc.id,
                "score": float(score),
            }
            for doc, score in zip(payload.documents, scores)
        ]

        results = [
            RankedDocument(
                id=item["id"],
                score=item["score"],
                rank=index + 1,
            )
            for index, item in enumerate(scored_documents)
        ]

        return RerankResponse(
            query=payload.query,
            model=MODEL_NAME,
            results=results,
        )

    except Exception as exc:
        logger.exception("Failed to rerank documents")

        raise HTTPException(
            status_code=500,
            detail="Failed to rerank documents",
        ) from exc