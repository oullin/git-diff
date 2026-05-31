import { app, BrowserWindow } from 'electron';
import { join } from 'node:path';
import { appIcon } from '#electron/app-icon.js';
import { attachWindowDiagnostics } from '#electron/diagnostics.js';
import { electronDir, repoRoot } from '#electron/paths.js';
import type { LaunchIntent } from '#electron/launch-intent.js';

const appWindowWidth = 2000;
const appWindowHeight = 1280;
const appWindowMinWidth = 1120;
const appWindowMinHeight = 760;
const devToolsWindowWidth = 1100;
const devToolsWindowHeight = 800;
const devToolsWindowMinWidth = 600;
const devToolsWindowMinHeight = 400;
const cascadeOffset = 32;

interface WindowEntry {
	window: BrowserWindow;
	intent: LaunchIntent | null;
	devTools: BrowserWindow | null;
}

const windows = new Map<number, WindowEntry>();

export function getMainWindow(): BrowserWindow | null {
	const focused = BrowserWindow.getFocusedWindow();

	if (focused && !focused.isDestroyed() && windows.has(focused.webContents.id)) {
		return focused;
	}

	for (const entry of windows.values()) {
		if (!entry.window.isDestroyed()) {
			return entry.window;
		}
	}

	return null;
}

export function listAppWindows(): BrowserWindow[] {
	return Array.from(windows.values())
		.map((entry) => entry.window)
		.filter((window) => !window.isDestroyed());
}

export function createWindow(intent: LaunchIntent | null = null): BrowserWindow {
	const icon = appIcon();
	const existing = listAppWindows();
	const offset = existing.length * cascadeOffset;

	const window = new BrowserWindow({
		width: appWindowWidth,
		height: appWindowHeight,
		minWidth: appWindowMinWidth,
		minHeight: appWindowMinHeight,
		...(existing.length === 0
			? { center: true }
			: (() => {
					const base = existing[existing.length - 1]!.getBounds();

					return { x: base.x + offset, y: base.y + offset };
				})()),
		resizable: true,
		maximizable: true,
		fullscreenable: true,
		title: 'Git Diff Review',
		vibrancy: 'sidebar',
		visualEffectState: 'active',
		backgroundColor: '#00000000',
		...(icon ? { icon } : {}),
		webPreferences: {
			preload: join(electronDir, 'preload.cjs'),
			contextIsolation: true,
			nodeIntegration: false,
			additionalArguments: intent ? [`--git-diff-intent=${JSON.stringify(intent)}`] : [],
		},
	});

	attachWindowDiagnostics(window);

	const id = window.webContents.id;

	windows.set(id, { window, intent, devTools: null });

	window.on('closed', () => {
		const entry = windows.get(id);

		if (entry?.devTools && !entry.devTools.isDestroyed()) {
			entry.devTools.close();
		}

		windows.delete(id);
	});

	const devServer = process.env.VITE_DEV_SERVER_URL;

	if (devServer) {
		void window.loadURL(devServer);
	} else if (app.isPackaged) {
		void window.loadFile(join(electronDir, '..', 'dist', 'index.html'));
	} else {
		void window.loadFile(join(repoRoot, 'packages', 'ui', 'dist', 'index.html'));
	}

	return window;
}

export function focusMainWindow(): void {
	const window = getMainWindow();

	if (!window) {
		return;
	}

	if (window.isMinimized()) {
		window.restore();
	}

	window.focus();
}

export function takeIntentForWindow(window: BrowserWindow): LaunchIntent | null {
	const entry = windows.get(window.webContents.id);

	if (!entry) {
		return null;
	}

	const intent = entry.intent;

	entry.intent = null;

	return intent;
}

export function setIntentForWindow(window: BrowserWindow, intent: LaunchIntent | null): void {
	const entry = windows.get(window.webContents.id);

	if (!entry) {
		return;
	}

	entry.intent = intent;
}

export function openDevToolsPanel(parentWindow: BrowserWindow): void {
	const entry = windows.get(parentWindow.webContents.id);

	if (entry?.devTools && !entry.devTools.isDestroyed()) {
		entry.devTools.close();
	}

	const devToolsWindow = new BrowserWindow({
		width: devToolsWindowWidth,
		height: devToolsWindowHeight,
		minWidth: devToolsWindowMinWidth,
		minHeight: devToolsWindowMinHeight,
		resizable: true,
		show: false,
		title: 'Git Diff Review DevTools',
		icon: appIcon(),
	});

	if (entry) {
		entry.devTools = devToolsWindow;
	}

	devToolsWindow.on('closed', () => {
		if (entry) {
			entry.devTools = null;
		}
	});

	devToolsWindow.once('ready-to-show', () => {
		const parentBounds = parentWindow.getBounds();

		devToolsWindow.setSize(devToolsWindowWidth, devToolsWindowHeight, false);
		devToolsWindow.setBounds({
			x: parentBounds.x + parentBounds.width,
			y: parentBounds.y,
			width: devToolsWindowWidth,
			height: devToolsWindowHeight,
		});
		devToolsWindow.show();
	});

	parentWindow.webContents.setDevToolsWebContents(devToolsWindow.webContents);
	parentWindow.webContents.openDevTools({ mode: 'detach' });
}
