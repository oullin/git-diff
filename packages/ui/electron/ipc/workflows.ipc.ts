import { ipcMain } from "electron";
import type { RunWorkflowRequest, WorkflowEvent } from "@git-diff/contracts";
import { client } from "#electron/bridge.js";

export function register(): void {
  ipcMain.handle("workflows:list", async () => {
    const response = await (await client()).listWorkflows();

    return response.workflows ?? [];
  });

  ipcMain.handle("runs:list", async (_event, limit: number) => {
    const response = await (await client()).listRuns({ limit });

    return response.runs ?? [];
  });

  ipcMain.handle("runs:log", async (_event, runId: string) => (await client()).runLog({ runId }));

  ipcMain.handle("template-files:list", async () => {
    const response = await (await client()).listTemplateFiles();

    return response.files ?? [];
  });

  ipcMain.handle("template-files:read", async (_event, path: string) =>
    (await client()).readTemplateFile({ path }),
  );

  ipcMain.handle("template-files:save", async (_event, path: string, content: string) =>
    (await client()).saveTemplateFile({ path, content }),
  );

  ipcMain.handle(
    "workflow:run",
    async (event, request: RunWorkflowRequest, eventChannel: string) => {
      const c = await client();

      return new Promise<{ exitCode: number }>((resolveResult, reject) => {
        const stream = c.runWorkflow(request);
        let exitCode = 0;

        stream.on("data", (workflowEvent: WorkflowEvent) => {
          if (workflowEvent.type === "run_failed") {
            exitCode = 1;
          }

          event.sender.send(eventChannel, workflowEvent);
        });

        stream.on("error", reject);
        stream.on("end", () => resolveResult({ exitCode }));
      });
    },
  );
}
