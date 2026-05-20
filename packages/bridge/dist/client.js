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
  getUIPreferences() {
    return this.request("GET", "/v1/preferences");
  }
  saveUIPreferences(values) {
    return this.request("POST", "/v1/preferences", { values });
  }
  getAuthState() {
    return this.request("GET", "/v1/auth/state");
  }
  authSetup(request) {
    return this.request("POST", "/v1/auth/setup", {
      password: request.password,
    });
  }
  authLogin(request) {
    return this.request("POST", "/v1/auth/login", {
      password: request.password,
      remember: request.remember,
    });
  }
  authResume(request) {
    return this.request("POST", "/v1/auth/resume", { token: request.token });
  }
  authLogout() {
    return this.request("POST", "/v1/auth/logout");
  }
  authWipe(request) {
    return this.request("POST", "/v1/auth/wipe", { osUsername: request.osUsername ?? "" });
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
  readCommit(request) {
    const parts = [`sha=${encodeURIComponent(request.sha)}`];
    if (request.path) {
      parts.push(`path=${encodeURIComponent(request.path)}`);
    }
    return this.request("GET", `/v1/repository/commit?${parts.join("&")}`);
  }
  listCommits(request) {
    const parts = [];
    if (request.path) {
      parts.push(`path=${encodeURIComponent(request.path)}`);
    }
    if (request.limit) {
      parts.push(`limit=${request.limit}`);
    }
    const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
    return this.request("GET", `/v1/repository/log${query}`);
  }
  readRepositoryFile(request) {
    const query = `?root=${encodeURIComponent(request.root)}&path=${encodeURIComponent(request.path)}`;
    return this.request("GET", `/v1/repository/file${query}`);
  }
  listBranches(request) {
    const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
    return this.request("GET", `/v1/repository/branches${query}`);
  }
  checkoutBranch(request) {
    return this.request("POST", "/v1/repository/checkout", {
      path: request.path,
      branch: request.branch,
    });
  }
  createBranch(request) {
    return this.request("POST", "/v1/repository/branches/create", {
      path: request.path,
      name: request.name,
    });
  }
  deleteBranch(request) {
    const params = new URLSearchParams({ name: request.name });
    if (request.path) params.set("path", request.path);
    return this.request("DELETE", `/v1/repository/branches?${params.toString()}`);
  }
  lockBranch(request) {
    return this.request("POST", "/v1/repository/branches/lock", {
      path: request.path,
      name: request.name,
    });
  }
  unlockBranch(request) {
    return this.request("POST", "/v1/repository/branches/unlock", {
      path: request.path,
      name: request.name,
    });
  }
  getSystemStats() {
    return this.request("GET", "/v1/system/stats");
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
  listRepositories() {
    return this.request("GET", "/v1/repositories");
  }
  searchRepositoryFiles(request) {
    const params = new URLSearchParams({ q: request.query });
    if (typeof request.limit === "number") {
      params.set("limit", String(request.limit));
    }
    return this.request("GET", `/v1/repositories/search-files?${params.toString()}`);
  }
  upsertRepository(request) {
    return this.request("POST", "/v1/repositories", {
      path: request.path,
      name: request.name ?? "",
    });
  }
  removeRepository(request) {
    return this.request("DELETE", `/v1/repositories?path=${encodeURIComponent(request.path)}`);
  }
  listCollaborators(request) {
    return this.request(
      "GET",
      `/v1/repositories/collaborators?path=${encodeURIComponent(request.path)}`,
    );
  }
  addCollaborator(request) {
    return this.request("POST", "/v1/repositories/collaborators", {
      path: request.path,
      userId: request.userId,
      role: request.role,
    });
  }
  removeCollaborator(request) {
    const params = new URLSearchParams({
      path: request.path,
      userId: String(request.userId),
    });
    return this.request("DELETE", `/v1/repositories/collaborators?${params.toString()}`);
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
