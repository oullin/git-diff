import type { RunLog, RunSummary, RunWorkflowRequest, TemplateFileContent, TemplateFileSummary, Workflow } from "@git-diff/contracts";
import type { WorkflowRunStream } from "#bridge/client-types.js";
export declare function listWorkflows(socketPath: string): Promise<{
    workflows: Workflow[];
}>;
export declare function listRuns(socketPath: string, request?: {
    limit?: number;
}): Promise<{
    runs: RunSummary[];
}>;
export declare function runLog(socketPath: string, request: {
    runId: string;
}): Promise<RunLog>;
export declare function listTemplateFiles(socketPath: string): Promise<{
    files: TemplateFileSummary[];
}>;
export declare function readTemplateFile(socketPath: string, request: {
    path: string;
}): Promise<TemplateFileContent>;
export declare function saveTemplateFile(socketPath: string, request: {
    path: string;
    content: string;
}): Promise<TemplateFileContent>;
export declare function runWorkflow(socketPath: string, request: RunWorkflowRequest): WorkflowRunStream;
