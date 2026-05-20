import type {
  RunLog,
  RunSummary,
  RunWorkflowRequest,
  TemplateFileContent,
  TemplateFileSummary,
  Workflow,
} from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";
import { runWorkflowStream } from "#bridge/sse.js";
import type { WorkflowRunStream } from "#bridge/client-types.js";

export function listWorkflows(socketPath: string): Promise<{ workflows: Workflow[] }> {
  return requestJson<{ workflows: Workflow[] }>(socketPath, "GET", "/v1/workflows");
}

export function listRuns(
  socketPath: string,
  request: { limit?: number } = {},
): Promise<{ runs: RunSummary[] }> {
  const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";
  return requestJson<{ runs: RunSummary[] }>(socketPath, "GET", `/v1/runs${query}`);
}

export function runLog(socketPath: string, request: { runId: string }): Promise<RunLog> {
  return requestJson<RunLog>(
    socketPath,
    "GET",
    `/v1/runs/${encodeURIComponent(request.runId)}/log`,
  );
}

export function listTemplateFiles(socketPath: string): Promise<{ files: TemplateFileSummary[] }> {
  return requestJson<{ files: TemplateFileSummary[] }>(socketPath, "GET", "/v1/template-files");
}

export function readTemplateFile(
  socketPath: string,
  request: { path: string },
): Promise<TemplateFileContent> {
  return requestJson<TemplateFileContent>(
    socketPath,
    "GET",
    `/v1/template-files/content?path=${encodeURIComponent(request.path)}`,
  );
}

export function saveTemplateFile(
  socketPath: string,
  request: { path: string; content: string },
): Promise<TemplateFileContent> {
  return requestJson<TemplateFileContent>(socketPath, "PUT", "/v1/template-files/content", {
    path: request.path,
    content: request.content,
  });
}

export function runWorkflow(socketPath: string, request: RunWorkflowRequest): WorkflowRunStream {
  return runWorkflowStream(socketPath, request);
}
