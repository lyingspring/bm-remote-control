import './tailwind.css';

import {SaveSSHConfig, LoadSSHConfig, ExecuteSSHCommand, TestSSHConnection, GetRemoteSystemInfo} from '/wailsjs/go/main/App';

let sshConfigModal = document.getElementById('sshConfigModal');
let toastContainer = document.getElementById('toastContainer');
let confirmDialog = document.getElementById('confirmDialog');
let confirmCallback = null;

// 连接状态监测
let heartbeatInterval = null;

// 自定义确认对话框
function showConfirm(title, message) {
    return new Promise((resolve) => {
        document.getElementById('confirmTitle').textContent = title;
        document.getElementById('confirmMessage').textContent = message;
        confirmCallback = resolve;
        confirmDialog.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
    });
}

function confirmDialogOk() {
    confirmDialog.classList.add('hidden');
    document.body.style.overflow = '';
    if (confirmCallback) {
        confirmCallback(true);
        confirmCallback = null;
    }
}

function confirmDialogCancel() {
    confirmDialog.classList.add('hidden');
    document.body.style.overflow = '';
    if (confirmCallback) {
        confirmCallback(false);
        confirmCallback = null;
    }
}

window.confirmDialogOk = confirmDialogOk;
window.confirmDialogCancel = confirmDialogCancel;

// 显示提示框
function showToast(message, type = 'info') {
    const toast = document.createElement('div');

    // 根据类型设置颜色和图标
    const config = {
        success: {
            color: '#34d399',
            bgColor: 'linear-gradient(135deg, rgba(52, 211, 153, 0.15) 0%, rgba(52, 211, 153, 0.05) 100%)',
            borderColor: 'rgba(52, 211, 153, 0.3)',
            icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>'
        },
        error: {
            color: '#f87171',
            bgColor: 'linear-gradient(135deg, rgba(248, 113, 113, 0.15) 0%, rgba(248, 113, 113, 0.05) 100%)',
            borderColor: 'rgba(248, 113, 113, 0.3)',
            icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>'
        },
        info: {
            color: '#60a5fa',
            bgColor: 'linear-gradient(135deg, rgba(96, 165, 250, 0.15) 0%, rgba(96, 165, 250, 0.05) 100%)',
            borderColor: 'rgba(96, 165, 250, 0.3)',
            icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>'
        },
        warning: {
            color: '#fbbf24',
            bgColor: 'linear-gradient(135deg, rgba(251, 191, 36, 0.15) 0%, rgba(251, 191, 36, 0.05) 100%)',
            borderColor: 'rgba(251, 191, 36, 0.3)',
            icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>'
        }
    };

    const cfg = config[type] || config.info;

    toast.className = 'toast pointer-events-auto';
    toast.style.cssText = `
        background: ${cfg.bgColor};
        border: 1px solid ${cfg.borderColor};
        color: ${cfg.color};
        padding: 1rem 1.25rem;
        border-radius: 0.75rem;
        box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.5);
        display: flex;
        align-items: center;
        gap: 0.75rem;
        min-width: 320px;
        max-width: 420px;
        font-size: 0.875rem;
        font-weight: 500;
    `;
    toast.innerHTML = `
        <div style="color: ${cfg.color}; flex-shrink: 0;">${cfg.icon}</div>
        <span class="flex-1" style="color: var(--color-text-primary);">${message}</span>
    `;

    toastContainer.appendChild(toast);

    // 3秒后自动移除（动画会自动淡出）
    setTimeout(() => {
        if (toast.parentNode) {
            toast.parentNode.removeChild(toast);
        }
    }, 3000);
}

// 打开 SSH 配置模态框
window.openSSHConfigModal = function() {
    sshConfigModal.classList.remove('hidden');
    document.body.style.overflow = 'hidden'; // 防止背景滚动
};

// 关闭 SSH 配置模态框
window.closeSSHConfigModal = function() {
    sshConfigModal.classList.add('hidden');
    document.body.style.overflow = ''; // 恢复背景滚动
};

// 更新连接状态指示器
function updateConnectionStatus(status) {
    const dot = document.getElementById('connectionStatusDot');
    const text = document.getElementById('connectionStatusText');

    // 重置状态
    dot.style.color = '';
    dot.style.background = '';

    switch (status) {
        case 'connected':
            dot.style.background = '#34d399';
            dot.style.color = '#34d399';
            text.textContent = '已连接';
            text.style.color = '#34d399';
            break;
        case 'disconnected':
            dot.style.background = '#f87171';
            dot.style.color = '#f87171';
            text.textContent = '未连接';
            text.style.color = '#f87171';
            break;
        case 'checking':
            dot.style.background = '#fbbf24';
            dot.style.color = '#fbbf24';
            text.textContent = '检查中...';
            text.style.color = '#fbbf24';
            break;
        default:
            dot.style.background = '#707080';
            dot.style.color = '#707080';
            text.textContent = '未知';
            text.style.color = '#707080';
    }
}

// 检查连接状态（心跳检测）
function checkConnectionStatus() {
    updateConnectionStatus('checking');

    // 使用轻量级的 echo 命令进行心跳检测
    ExecuteSSHCommand('echo "heartbeat"')
        .then((output) => {
            if (output && output.includes('heartbeat')) {
                updateConnectionStatus('connected');
            } else {
                updateConnectionStatus('disconnected');
            }
        })
        .catch((err) => {
            updateConnectionStatus('disconnected');
        });
}

// 启动心跳监测
function startHeartbeat() {
    // 立即执行一次检查
    checkConnectionStatus();

    // 每 30 秒检查一次
    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
    }
    heartbeatInterval = setInterval(checkConnectionStatus, 30000);
}

// 页面加载时执行
window.onload = function() {
    loadSSHConfig();
    loadRemoteSystemInfo();
    startHeartbeat(); // 启动心跳监测

    // ESC 键关闭模态框
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && !sshConfigModal.classList.contains('hidden')) {
            window.closeSSHConfigModal();
        }
    });
};

// 加载远程主机信息
function loadRemoteSystemInfo() {
    GetRemoteSystemInfo()
        .then((info) => {
            document.getElementById('remoteHost').textContent = info.hostname || '未知';
            document.getElementById('remoteOS').textContent = info.os || '未知';
            document.getElementById('remoteArch').textContent = info.arch || '未知';
            document.getElementById('remoteUptime').textContent = info.uptime || '未知';
        })
        .catch(() => {
            console.log('尚未连接远程主机');
        });
}

// 加载 SSH 配置
function loadSSHConfig() {
    LoadSSHConfig()
        .then((config) => {
            document.getElementById('sshHost').value = config.host || '';
            document.getElementById('sshPort').value = config.port || '22';
            document.getElementById('sshUsername').value = config.username || '';
            document.getElementById('sshPassword').value = config.password || '';
        })
        .catch((err) => {
            console.error('加载 SSH 配置失败:', err);
        });
}

// 保存 SSH 配置
window.saveSSHConfig = function() {
    const host = document.getElementById('sshHost').value;
    const port = document.getElementById('sshPort').value || '22';
    const username = document.getElementById('sshUsername').value;
    const password = document.getElementById('sshPassword').value;

    if (!host || !username) {
        showToast('请至少填写 IP 地址和用户名', 'warning');
        return;
    }

    SaveSSHConfig(host, port, username, password)
        .then((result) => {
            showToast('SSH 配置已成功保存', 'success');
            // 立即检查连接状态
            checkConnectionStatus();
            // 刷新远程主机信息
            loadRemoteSystemInfo();
        })
        .catch((err) => {
            showToast('保存失败：' + err, 'error');
        });
};

// 测试 SSH 连接
window.testSSHConnection = function() {
    showToast('正在测试连接...', 'info');

    TestSSHConnection()
        .then((result) => {
            showToast('SSH 连接测试成功！', 'success');
            updateConnectionStatus('connected');
            // 刷新远程主机信息
            loadRemoteSystemInfo();
        })
        .catch((err) => {
            showToast('连接测试失败：' + err, 'error');
            updateConnectionStatus('disconnected');
        });
};

// 重启服务
window.restartService = async function(serviceFile, serviceName) {
    const confirmed = await showConfirm('重启服务', `确定要重启服务 ${serviceName} 吗？`);
    if (!confirmed) {
        return;
    }

    showToast(`正在重启服务 ${serviceName}...`, 'info');

    try {
        const command = `sudo systemctl restart ${serviceFile}`;
        const output = await ExecuteSSHCommand(command);

        if (!output) {
            showToast(`${serviceName} 重启成功！`, 'success');
        } else {
            showToast(`${serviceName} 重启完成`, 'success');
        }
    } catch (err) {
        showToast(`重启失败：${err}`, 'error');
    }
};

// 远程关机
window.remoteShutdown = async function() {
    const confirmed = await showConfirm('远程关机', '确定要关闭远程主机吗？此操作将立即关闭远程主机。');
    if (!confirmed) {
        return;
    }

    showToast('正在发送关机命令...', 'info');

    // 尝试多种关机命令（优先尝试 sudo 命令，因为会自动输入密码）
    const shutdownCommands = [
        'sudo systemctl poweroff',      // 使用 sudo 的 systemd 关机（优先）
        'sudo shutdown -h now',         // 使用 sudo 的传统关机
        'sudo poweroff',                // 使用 sudo 直接关机
        'systemctl poweroff',          // 不使用 sudo
        'shutdown -h now',             // 不使用 sudo
        'poweroff'                     // 不使用 sudo
    ];

    tryCommand(shutdownCommands, '关机');
};

// 远程重启
window.remoteRestart = async function() {
    const confirmed = await showConfirm('远程重启', '确定要重启远程主机吗？此操作将立即重启远程主机。');
    if (!confirmed) {
        return;
    }

    showToast('正在发送重启命令...', 'info');

    // 尝试多种重启命令（优先尝试 sudo 命令，因为会自动输入密码）
    const restartCommands = [
        'sudo systemctl reboot',        // 使用 sudo 的 systemd 重启（优先）
        'sudo shutdown -r now',         // 使用 sudo 的传统重启
        'sudo reboot',                  // 使用 sudo 直接重启
        'systemctl reboot',             // 不使用 sudo
        'shutdown -r now',             // 不使用 sudo
        'reboot'                       // 不使用 sudo
    ];

    tryCommand(restartCommands, '重启');
};

// 尝试执行多个命令，直到成功
async function tryCommand(commands, operationName) {
    let lastError = '';

    for (const cmd of commands) {
        try {
            await ExecuteSSHCommand(cmd);
            showToast(`${operationName}命令已成功发送`, 'success');
            return;
        } catch (err) {
            lastError = err.toString();
            continue; // 尝试下一个命令
        }
    }

    // 所有命令都失败
    showToast(`${operationName}失败：` + lastError, 'error');
}
