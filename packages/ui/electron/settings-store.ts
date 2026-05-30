import type { RuntimeSettings } from "@git-diff/contracts";
import { app } from "electron";
import {
    copyFileSync,
    existsSync,
    mkdirSync,
    readFileSync,
    statSync,
    unlinkSync,
    writeFileSync,
} from "node:fs";
import { dirname, join } from "node:path";

export interface SettingsStore {
    read(): Partial<RuntimeSettings>;
    write(settings: Partial<RuntimeSettings>): void;
    clean(settings: Partial<RuntimeSettings>): Partial<RuntimeSettings>;
    settingsPath(): string;
    defaultWorkflowDbPath(): string;
}

class ElectronSettingsStore implements SettingsStore {
    settingsPath(): string {
        const devSettingsPath = process.env.GIT_DIFF_SETTINGS_PATH?.trim();

        if (devSettingsPath) {
            return devSettingsPath;
        }

        return join(app.getPath("userData"), "settings.json");
    }

    defaultWorkflowDbPath(): string {
        return join(app.getPath("userData"), "reviews.sqlite3");
    }

    read(): Partial<RuntimeSettings> {
        try {
            const raw = JSON.parse(
                readFileSync(this.settingsPath(), "utf8"),
            ) as Partial<RuntimeSettings>;

            return this.clean(raw);
        } catch {
            return this.clean({});
        }
    }

    write(settings: Partial<RuntimeSettings>): void {
        mkdirSync(dirname(this.settingsPath()), { recursive: true });
        writeFileSync(
            this.settingsPath(),
            JSON.stringify(this.clean(settings), null, 2) + "\n",
            "utf8",
        );
    }

    clean(settings: Partial<RuntimeSettings>): Partial<RuntimeSettings> {
        return {
            repoRoot: settings.repoRoot ?? "",
            appsConfigPath: settings.appsConfigPath ?? "",
            secretsConfigPath: settings.secretsConfigPath ?? "",
            generatedAppsPath: settings.generatedAppsPath ?? "",
            archiveRoot: settings.archiveRoot ?? "",
            workflowDbPath: settings.workflowDbPath?.trim()
                ? settings.workflowDbPath
                : this.defaultWorkflowDbPath(),
            opVault: settings.opVault ?? "",
            opItem: settings.opItem ?? "",
        };
    }
}

let activeStore: SettingsStore = new ElectronSettingsStore();

/** Returns a restore func for scoped overrides. */
export function setSettingsStore(store: SettingsStore): () => void {
    const prev = activeStore;

    activeStore = store;

    return () => {
        activeStore = prev;
    };
}

export function settingsPath(): string {
    return activeStore.settingsPath();
}

export function defaultWorkflowDbPath(): string {
    return activeStore.defaultWorkflowDbPath();
}

export function readSavedSettings(): Partial<RuntimeSettings> {
    return activeStore.read();
}

export function writeSavedSettings(settings: Partial<RuntimeSettings>): void {
    activeStore.write(settings);
}

export function cleanSettings(settings: Partial<RuntimeSettings>): Partial<RuntimeSettings> {
    return activeStore.clean(settings);
}

export function moveWorkflowDatabase(fromPath?: string, toPath?: string): () => void {
    if (!fromPath || !toPath || fromPath === toPath) {
        return () => {};
    }

    if (existsSync(toPath) && statSync(toPath).isDirectory()) {
        throw new Error(`Review database path is a directory: ${toPath}`);
    }

    mkdirSync(dirname(toPath), { recursive: true });

    if (!existsSync(fromPath)) {
        return () => {};
    }

    if (existsSync(toPath)) {
        throw new Error(`Review database already exists: ${toPath}`);
    }

    copyFileSync(fromPath, toPath);
    unlinkSync(fromPath);

    return () => {
        if (existsSync(toPath) && !existsSync(fromPath)) {
            mkdirSync(dirname(fromPath), { recursive: true });
            copyFileSync(toPath, fromPath);
            unlinkSync(toPath);
        }
    };
}
