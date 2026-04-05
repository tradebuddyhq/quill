const vscode = require('vscode');
const { exec } = require('child_process');
const path = require('path');

function activate(context) {
    // Run command
    let runCmd = vscode.commands.registerCommand('quill.run', () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) return;
        const file = editor.document.fileName;
        if (!file.endsWith('.quill')) {
            vscode.window.showErrorMessage('Not a .quill file');
            return;
        }

        // Save first
        editor.document.save().then(() => {
            const terminal = vscode.window.createTerminal('Quill');
            terminal.show();
            terminal.sendText(`quill run "${file}"`);
        });
    });

    // Build command
    let buildCmd = vscode.commands.registerCommand('quill.build', () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) return;
        const file = editor.document.fileName;
        if (!file.endsWith('.quill')) return;

        editor.document.save().then(() => {
            const terminal = vscode.window.createTerminal('Quill');
            terminal.show();
            terminal.sendText(`quill build "${file}"`);
        });
    });

    // Init command
    let initCmd = vscode.commands.registerCommand('quill.init', () => {
        const folder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
        if (!folder) {
            vscode.window.showErrorMessage('Open a folder first');
            return;
        }
        const terminal = vscode.window.createTerminal('Quill');
        terminal.show();
        terminal.sendText(`cd "${folder}" && quill init`);
    });

    context.subscriptions.push(runCmd, buildCmd, initCmd);

    // Status bar item
    const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBar.text = '$(play) Run Quill';
    statusBar.command = 'quill.run';
    statusBar.tooltip = 'Run current Quill file (Ctrl+Shift+R)';
    context.subscriptions.push(statusBar);

    // Show status bar for .quill files
    vscode.window.onDidChangeActiveTextEditor(editor => {
        if (editor && editor.document.languageId === 'quill') {
            statusBar.show();
        } else {
            statusBar.hide();
        }
    });

    // Check on activation
    if (vscode.window.activeTextEditor?.document.languageId === 'quill') {
        statusBar.show();
    }
}

function deactivate() {}

module.exports = { activate, deactivate };
