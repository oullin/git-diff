export interface RuntimeSettings {
    repoRoot: string;
    appsConfigPath: string;
    secretsConfigPath: string;
    generatedAppsPath: string;
    archiveRoot: string;
    workflowDbPath: string;
    opVault: string;
    opItem: string;
}
export interface SettingsCheck {
    key: string;
    label: string;
    path: string;
    status: string;
    message: string;
}
export interface SettingsResponse {
    settings?: RuntimeSettings;
    checks: SettingsCheck[];
    valid: boolean;
}
//# sourceMappingURL=index.d.ts.map
