import { app } from 'electron';
import { parseLaunchArgs } from '#electron/launch-intent.js';
import { runApp, showHelpAndExit } from '#electron/lifecycle.js';

const initialIntent = parseLaunchArgs(process.argv, app.isPackaged, process.cwd());

if (initialIntent.kind === 'help') {
	showHelpAndExit(initialIntent.helpText ?? '');
} else {
	runApp(initialIntent);
}
