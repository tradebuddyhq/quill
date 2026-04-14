const vscode = require('vscode');
const { LanguageClient, TransportKind } = require('vscode-languageclient/node');

let client;

function activate(context) {
    // ---- LSP Client ----
    const serverOptions = {
        command: 'quill',
        args: ['lsp'],
        transport: TransportKind.stdio
    };

    const clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'quill' }],
    };

    client = new LanguageClient(
        'quillLanguageServer',
        'Quill Language Server',
        serverOptions,
        clientOptions
    );

    client.start();

    // ---- Run command ----
    let runCmd = vscode.commands.registerCommand('quill.run', () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) return;
        const file = editor.document.fileName;
        if (!file.endsWith('.quill')) {
            vscode.window.showErrorMessage('Not a .quill file');
            return;
        }

        editor.document.save().then(() => {
            const terminal = vscode.window.createTerminal('Quill');
            terminal.show();
            terminal.sendText(`quill run "${file}"`);
        });
    });

    // ---- Build command ----
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

    // ---- Init command ----
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

    // ---- Status bar ----
    const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBar.text = '$(play) Run Quill';
    statusBar.command = 'quill.run';
    statusBar.tooltip = 'Run current Quill file (Ctrl+Shift+R)';
    context.subscriptions.push(statusBar);

    vscode.window.onDidChangeActiveTextEditor(editor => {
        if (editor && editor.document.languageId === 'quill') {
            statusBar.show();
        } else {
            statusBar.hide();
        }
    });

    if (vscode.window.activeTextEditor?.document.languageId === 'quill') {
        statusBar.show();
    }
}

function deactivate() {
    if (client) {
        return client.stop();
    }
}

module.exports = { activate, deactivate };
