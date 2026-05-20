export interface Phase {
    id: string;
    name: string;
    enabled: boolean;
}
export interface ConfirmationOption {
    id: string;
    label: string;
    description: string;
    continue: boolean;
    back: boolean;
    requiresApproval: boolean;
    phases: Phase[];
}
export interface Workflow {
    id: string;
    name: string;
    description: string;
    changesMac: string;
    phases: Phase[];
    confirmation?: {
        title: string;
        message: string;
        options: ConfirmationOption[];
    };
}
export interface RunWorkflowRequest {
    workflowId: string;
    confirmationOptionId: string;
    enabledPhaseIds: string[];
}
export interface WorkflowEvent {
    id?: number;
    runId: string;
    seq: number;
    type: string;
    phaseId?: string;
    phaseName?: string;
    status?: string;
    message?: string;
    createdAt?: string;
}
export interface RunSummary {
    id: string;
    workflowId: string;
    workflowName: string;
    confirmationOptionId: string;
    confirmationOptionLabel: string;
    mode: string;
    status: string;
    startedAt: string;
    completedAt?: string;
    errorMessage?: string;
}
export interface RunLog {
    run?: RunSummary;
    events: WorkflowEvent[];
}
export interface TemplateFileSummary {
    path: string;
    relative: string;
    kind: string;
    size: number;
    modifiedAt?: string;
    exists: boolean;
}
export interface TemplateFileContent {
    file: TemplateFileSummary;
    content: string;
}
export interface WorkflowRunEndInfo {
    exitCode: number;
    status: string;
    message?: string;
}
//# sourceMappingURL=index.d.ts.map