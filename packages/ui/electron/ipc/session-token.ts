import { app, safeStorage } from "electron";
import fs from "node:fs";
import path from "node:path";

function sessionTokenPath(): string {
    return path.join(app.getPath("userData"), ".session-token");
}

export function readSessionToken(): string | null {
    const filePath = sessionTokenPath();

    if (!fs.existsSync(filePath)) {
        return null;
    }

    try {
        const buf = fs.readFileSync(filePath);

        if (safeStorage.isEncryptionAvailable()) {
            return safeStorage.decryptString(buf);
        }

        return buf.toString("utf8");
    } catch {
        try {
            fs.unlinkSync(filePath);
        } catch {
            /* ignore */
        }

        return null;
    }
}

export function writeSessionToken(token: string): void {
    const filePath = sessionTokenPath();

    try {
        fs.mkdirSync(path.dirname(filePath), { recursive: true });

        if (safeStorage.isEncryptionAvailable()) {
            fs.writeFileSync(filePath, safeStorage.encryptString(token), { mode: 0o600 });
        } else {
            console.warn("electron safeStorage unavailable; persisting session token in plaintext");
            fs.writeFileSync(filePath, token, { mode: 0o600, encoding: "utf8" });
        }
    } catch (error) {
        console.warn("failed to persist session token", error);
    }
}

export function clearSessionToken(): void {
    const filePath = sessionTokenPath();

    try {
        if (fs.existsSync(filePath)) {
            fs.unlinkSync(filePath);
        }
    } catch {
        /* ignore */
    }
}
