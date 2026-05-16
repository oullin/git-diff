import { requestJson } from "#bridge/http.js";
import { runWorkflowStream } from "#bridge/sse.js";
export function unixTarget(socketPath) {
  return { socketPath };
}
export function createWorkflowBridgeClient(target) {
  return new HttpWorkflowBridgeClient(target);
}
export function waitForReady(client, timeoutMs = 10000) {
  const deadline = Date.now() + timeoutMs;
  return new Promise((resolve, reject) => {
    const attempt = () => {
      client
        .healthz()
        .then(resolve)
        .catch((error) => {
          if (Date.now() >= deadline) {
            reject(error);
            return;
          }
          setTimeout(attempt, 100);
        });
    };
    attempt();
  });
}
class HttpWorkflowBridgeClient {
  socketPath;
  constructor(target) {
    this.socketPath = target.socketPath;
  }
  close() {}
  healthz() {
    return this.request("GET", "/v1/healthz");
  }
  listWorkflows() {
    return this.request("GET", "/v1/workflows");
  }
  listRuns(request = {}) {
    const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";
    return this.request("GET", `/v1/runs${query}`);
  }
  runLog(request) {
    return this.request("GET", `/v1/runs/${encodeURIComponent(request.runId)}/log`);
  }
  listTemplateFiles() {
    return this.request("GET", "/v1/template-files");
  }
  readTemplateFile(request) {
    return this.request(
      "GET",
      `/v1/template-files/content?path=${encodeURIComponent(request.path)}`,
    );
  }
  saveTemplateFile(request) {
    return this.request("PUT", "/v1/template-files/content", {
      path: request.path,
      content: request.content,
    });
  }
  getSettings() {
    return this.request("GET", "/v1/settings");
  }
  validateSettings(request) {
    return this.request("POST", "/v1/settings/validate", {
      settings: request.settings,
    });
  }
  getUserPreferences() {
    return this.request("GET", "/v1/preferences");
  }
  saveUserPreferences(request) {
    return this.request("POST", "/v1/preferences", {
      theme: request.theme,
      diffViewMode: request.diffViewMode,
      hideWhitespace: request.hideWhitespace,
      lastRepoRoot: request.lastRepoRoot,
    });
  }
  repositoryState(request) {
    const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
    return this.request("GET", `/v1/repository/state${query}`);
  }
  openRepository(request) {
    return this.request("POST", "/v1/repository/open", { path: request.path });
  }
  refreshRepository(request) {
    return this.request("POST", "/v1/repository/refresh", { path: request.path });
  }
  createReview(request) {
    return this.request("POST", "/v1/reviews", request);
  }
  listReviews(request = {}) {
    const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";
    return this.request("GET", `/v1/reviews${query}`);
  }
  reviewDetail(request) {
    return this.request("GET", `/v1/reviews/${encodeURIComponent(request.id)}`);
  }
  addReviewEvent(request) {
    return this.request("POST", `/v1/reviews/${encodeURIComponent(request.reviewId)}/events`, {
      type: request.type,
      filePath: request.filePath,
      message: request.message,
      metadata: request.metadata,
    });
  }
  createReviewComment(request) {
    return this.request("POST", `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments`, {
      filePath: request.filePath,
      diffSection: request.diffSection,
      side: request.side,
      lineNumber: request.lineNumber,
      authorLabel: request.authorLabel,
      bodyHtml: request.bodyHtml,
    });
  }
  updateReviewComment(request) {
    return this.request(
      "PATCH",
      `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
      { bodyHtml: request.bodyHtml },
    );
  }
  deleteReviewComment(request) {
    return this.request(
      "DELETE",
      `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
    );
  }
  listOpVaults() {
    return this.request("GET", "/v1/onepassword/vaults");
  }
  listOpItems(request) {
    return this.request("GET", `/v1/onepassword/items?vault=${encodeURIComponent(request.vault)}`);
  }
  runWorkflow(request) {
    return runWorkflowStream(this.socketPath, request);
  }
  request(method, path, body) {
    return requestJson(this.socketPath, method, path, body);
  }
}
