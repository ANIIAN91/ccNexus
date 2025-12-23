import { api } from '../api.js';
import { notifications } from '../utils/notifications.js';

class WebDAV {
    constructor() {
        this.container = document.getElementById('view-container');
        this.backups = [];
    }

    async render() {
        this.container.innerHTML = `
            <div class="webdav">
                <div class="flex-between mb-3">
                    <h1>WebDAV</h1>
                    <div class="flex gap-2">
                        <button class="btn btn-secondary" id="webdav-refresh">Refresh</button>
                        <button class="btn btn-primary" id="webdav-backup">Backup Now</button>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <h3 class="card-title">Configuration</h3>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="form-label">URL *</label>
                            <input class="form-input" id="webdav-url" placeholder="https://dav.example.com/remote.php/webdav" />
                        </div>
                        <div class="form-group">
                            <label class="form-label">Username</label>
                            <input class="form-input" id="webdav-username" placeholder="username" />
                        </div>
                        <div class="form-group">
                            <label class="form-label">Password</label>
                            <input class="form-input" id="webdav-password" type="password" placeholder="password" />
                            <small class="text-muted">Leave empty to keep existing password (if already configured)</small>
                        </div>
                        <div class="flex gap-2">
                            <button class="btn btn-secondary" id="webdav-test">Test</button>
                            <button class="btn btn-primary" id="webdav-save">Save</button>
                        </div>
                        <div id="webdav-test-result" class="mt-3" style="display:none;"></div>
                    </div>
                </div>

                <div class="card mt-3">
                    <div class="card-header">
                        <h3 class="card-title">Backups</h3>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="form-label">Backup filename (optional)</label>
                            <input class="form-input" id="webdav-filename" placeholder="backup-20250101-120000.db" />
                        </div>

                        <div id="webdav-backups"></div>

					<div id="webdav-conflict-result" class="mt-3" style="display:none;"></div>
                    </div>
                </div>
            </div>
        `;

        document.getElementById('webdav-refresh').addEventListener('click', () => this.refresh());
        document.getElementById('webdav-save').addEventListener('click', () => this.saveConfig());
        document.getElementById('webdav-test').addEventListener('click', () => this.testConnection());
        document.getElementById('webdav-backup').addEventListener('click', () => this.backupNow());

        await this.refresh();
    }

    async refresh() {
        await this.loadConfig();
        await this.loadBackups();
    }

    async loadConfig() {
        try {
            const cfg = await api.getWebDAVConfig();
            document.getElementById('webdav-url').value = cfg.url || '';
            document.getElementById('webdav-username').value = cfg.username || '';
            document.getElementById('webdav-password').value = '';
        } catch (error) {
            notifications.error('Failed to load WebDAV config: ' + error.message);
        }
    }

    async saveConfig() {
        const url = document.getElementById('webdav-url').value.trim();
        const username = document.getElementById('webdav-username').value.trim();
        const password = document.getElementById('webdav-password').value;

        if (!url) {
            notifications.warning('URL is required');
            return;
        }

        try {
            await api.updateWebDAVConfig(url, username, password);
            notifications.success('WebDAV config saved');
            document.getElementById('webdav-password').value = '';
        } catch (error) {
            notifications.error('Failed to save WebDAV config: ' + error.message);
        }
    }

    async testConnection() {
        const url = document.getElementById('webdav-url').value.trim();
        const username = document.getElementById('webdav-username').value.trim();
        const password = document.getElementById('webdav-password').value;

        if (!url) {
            notifications.warning('URL is required');
            return;
        }

        const resultDiv = document.getElementById('webdav-test-result');
        resultDiv.style.display = 'block';
        resultDiv.innerHTML = '<div class="flex-center"><div class="spinner"></div></div>';

        try {
            const result = await api.testWebDAV(url, username, password);
            const success = !!result.success;
            const message = this.escapeHtml(result.message || (success ? 'Success' : 'Failed'));

            resultDiv.innerHTML = `
                <div class="card" style="background-color: var(--bg-secondary);">
                    <div class="mb-2">
                        <span class="badge ${success ? 'badge-success' : 'badge-danger'}">${success ? 'Success' : 'Failed'}</span>
                    </div>
                    <div class="code-block">${message}</div>
                </div>
            `;

            if (success) {
                notifications.success('WebDAV test succeeded');
            } else {
                notifications.error('WebDAV test failed');
            }
        } catch (error) {
            resultDiv.innerHTML = `
                <div class="card" style="background-color: var(--bg-secondary);">
                    <div class="mb-2">
                        <span class="badge badge-danger">Error</span>
                    </div>
                    <div class="code-block">${this.escapeHtml(error.message)}</div>
                </div>
            `;
            notifications.error('WebDAV test failed: ' + error.message);
        }
    }

    async loadBackups() {
        const container = document.getElementById('webdav-backups');
        container.innerHTML = '<div class="flex-center"><div class="spinner"></div></div>';

        try {
            const result = await api.listWebDAVBackups();
            if (!result.success) {
                container.innerHTML = `<div class="empty-state"><p>${this.escapeHtml(result.message || 'Failed to load backups')}</p></div>`;
                return;
            }

            this.backups = Array.isArray(result.backups) ? result.backups : [];
            if (this.backups.length === 0) {
                container.innerHTML = '<div class="empty-state"><p>No backups found</p></div>';
                return;
            }

            container.innerHTML = `
                <div class="table-container">
                    <table class="table">
                        <thead>
                            <tr>
                                <th style="width: 40px;"></th>
                                <th>Filename</th>
                                <th style="width: 240px;">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${this.backups.map((b, idx) => {
                                const filename = this.escapeHtml(b.filename || b.name || '');
                                return `
                                    <tr>
                                        <td><input type="checkbox" class="webdav-backup-check" data-index="${idx}"></td>
                                        <td><code>${filename}</code></td>
                                        <td>
                                            <div class="flex gap-2">
                                                <button class="btn btn-sm btn-secondary" data-action="conflict" data-index="${idx}">Conflict</button>
                                                <button class="btn btn-sm btn-primary" data-action="restore" data-index="${idx}">Restore</button>
                                            </div>
                                        </td>
                                    </tr>
                                `;
                            }).join('')}
                        </tbody>
                    </table>
                </div>
                <div class="flex gap-2 mt-2">
                    <button class="btn btn-danger" id="webdav-delete">Delete Selected</button>
                </div>
            `;

            document.getElementById('webdav-delete').addEventListener('click', () => this.deleteSelected());
            container.querySelectorAll('button[data-action]').forEach(btn => {
                btn.addEventListener('click', () => {
                    const index = parseInt(btn.dataset.index);
                    const action = btn.dataset.action;
                    if (action === 'restore') this.restore(index);
                    if (action === 'conflict') this.conflict(index);
                });
            });
        } catch (error) {
            container.innerHTML = `<div class="empty-state"><p>${this.escapeHtml(error.message)}</p></div>`;
            notifications.error('Failed to load backups: ' + error.message);
        }
    }

    async backupNow() {
        const filename = document.getElementById('webdav-filename').value.trim();
        try {
            await api.backupToWebDAV(filename);
            notifications.success('Backup created');
            await this.loadBackups();
        } catch (error) {
            notifications.error('Backup failed: ' + error.message);
        }
    }

    async restore(index) {
        const backup = this.backups[index];
        const filename = (backup && (backup.filename || backup.name)) ? (backup.filename || backup.name) : '';
        if (!filename) {
            notifications.error('Invalid backup filename');
            return;
        }

        // Desktop-aligned behavior: check conflicts first; if conflicts exist, let user choose.
        const resultDiv = document.getElementById('webdav-conflict-result');
        resultDiv.style.display = 'block';
        resultDiv.innerHTML = '<div class="flex-center"><div class="spinner"></div></div>';

        try {
            const conflict = await api.detectWebDAVConflict(filename);
            const conflicts = Array.isArray(conflict.conflicts) ? conflict.conflicts : [];

            if (conflict.success && conflicts.length > 0) {
                this.renderRestoreChoice(filename, conflicts);
                notifications.warning('Conflict detected: choose restore strategy');
                return;
            }

            // Default (safer): merge while keeping local on conflicts.
            await api.restoreFromWebDAV(filename, 'local');
            notifications.success('Restore completed');
            resultDiv.style.display = 'none';
        } catch (error) {
            notifications.error('Restore failed: ' + error.message);
            resultDiv.innerHTML = `
                <div class="card" style="background-color: var(--bg-secondary);">
                    <div class="mb-2">
                        <span class="badge badge-danger">Restore Error</span>
                        <span class="text-muted ml-2">${this.escapeHtml(filename)}</span>
                    </div>
                    <div class="code-block">${this.escapeHtml(error.message)}</div>
                </div>
            `;
        }
    }

    renderRestoreChoice(filename, conflicts) {
        const resultDiv = document.getElementById('webdav-conflict-result');
        const conflictText = conflicts.length
            ? conflicts.map(c => this.escapeHtml(JSON.stringify(c))).join('<br/>')
            : 'No conflicts';

        resultDiv.style.display = 'block';
        resultDiv.innerHTML = `
            <div class="card" style="background-color: var(--bg-secondary);">
                <div class="mb-2">
                    <span class="badge badge-warning">Conflict Detected</span>
                    <span class="text-muted ml-2">${this.escapeHtml(filename)}</span>
                </div>
                <div class="mb-2">
                    <div class="code-block">${conflictText}</div>
                </div>
                <div class="flex gap-2">
                    <button class="btn btn-sm btn-primary" id="webdav-restore-remote">Use Remote (overwrite local)</button>
                    <button class="btn btn-sm btn-secondary" id="webdav-restore-local">Keep Local (merge)</button>
                    <button class="btn btn-sm btn-secondary" id="webdav-restore-cancel">Cancel</button>
                </div>
            </div>
        `;

        document.getElementById('webdav-restore-cancel').addEventListener('click', () => {
            resultDiv.style.display = 'none';
        });

        document.getElementById('webdav-restore-remote').addEventListener('click', async () => {
            try {
                await api.restoreFromWebDAV(filename, 'remote');
                notifications.success('Restore completed (remote overwrite)');
                resultDiv.style.display = 'none';
            } catch (error) {
                notifications.error('Restore failed: ' + error.message);
            }
        });

        document.getElementById('webdav-restore-local').addEventListener('click', async () => {
            try {
                await api.restoreFromWebDAV(filename, 'local');
                notifications.success('Restore completed (keep local)');
                resultDiv.style.display = 'none';
            } catch (error) {
                notifications.error('Restore failed: ' + error.message);
            }
        });
    }

    async conflict(index) {
        const backup = this.backups[index];
        const filename = (backup && (backup.filename || backup.name)) ? (backup.filename || backup.name) : '';
        if (!filename) {
            notifications.error('Invalid backup filename');
            return;
        }

        const resultDiv = document.getElementById('webdav-conflict-result');
        resultDiv.style.display = 'block';
        resultDiv.innerHTML = '<div class="flex-center"><div class="spinner"></div></div>';

        try {
            const result = await api.detectWebDAVConflict(filename);
            if (result.success) {
                const conflicts = Array.isArray(result.conflicts) ? result.conflicts : [];
                const text = conflicts.length
                    ? conflicts.map(c => this.escapeHtml(JSON.stringify(c))).join('<br/>')
                    : 'No conflicts';

                resultDiv.innerHTML = `
                    <div class="card" style="background-color: var(--bg-secondary);">
                        <div class="mb-2">
                            <span class="badge badge-info">Conflict Check</span>
                            <span class="text-muted ml-2">${this.escapeHtml(filename)}</span>
                        </div>
                        <div class="code-block">${text}</div>
                    </div>
                `;
                notifications.info('Conflict check done');
            } else {
                resultDiv.innerHTML = `
                    <div class="card" style="background-color: var(--bg-secondary);">
                        <div class="mb-2">
                            <span class="badge badge-danger">Failed</span>
                        </div>
                        <div class="code-block">${this.escapeHtml(result.message || 'Conflict check failed')}</div>
                    </div>
                `;
                notifications.error(result.message || 'Conflict check failed');
            }
        } catch (error) {
            resultDiv.innerHTML = `
                <div class="card" style="background-color: var(--bg-secondary);">
                    <div class="mb-2">
                        <span class="badge badge-danger">Error</span>
                    </div>
                    <div class="code-block">${this.escapeHtml(error.message)}</div>
                </div>
            `;
            notifications.error('Conflict check failed: ' + error.message);
        }
    }

    async deleteSelected() {
        const checks = Array.from(document.querySelectorAll('.webdav-backup-check'));
        const filenames = checks
            .filter(c => c.checked)
            .map(c => {
                const idx = parseInt(c.dataset.index);
                const b = this.backups[idx];
                return (b && (b.filename || b.name)) ? (b.filename || b.name) : '';
            })
            .filter(Boolean);

        if (filenames.length === 0) {
            notifications.warning('No backups selected');
            return;
        }

        try {
            await api.deleteWebDAVBackups(filenames);
            notifications.success('Deleted backups');
            await this.loadBackups();
        } catch (error) {
            notifications.error('Delete failed: ' + error.message);
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = String(text);
        return div.innerHTML;
    }
}

export const webdav = new WebDAV();
