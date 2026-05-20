import type { RunWorkflowRequest } from "@git-diff/contracts";
import type { WorkflowRunStream } from "#bridge/client-types.js";
export declare function runWorkflowStream(socketPath: string, request: RunWorkflowRequest): WorkflowRunStream;
