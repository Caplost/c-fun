// 全局变量
let currentUser = { id: 1, username: 'testuser' }; // 默认用户，实际应用中应使用登录用户

// DOM 元素引用
const problemsList = document.getElementById('problemsList');
const submissionsList = document.getElementById('submissionsList');
const refreshProblemsBtn = document.getElementById('refreshProblemsBtn');
const createProblemBtn = document.getElementById('createProblemBtn');

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    // 事件监听器
    if (refreshProblemsBtn) {
        refreshProblemsBtn.addEventListener('click', loadProblems);
    }
    
    if (createProblemBtn) {
        createProblemBtn.addEventListener('click', showCreateProblemModal);
    }
    
    // 添加导入题目的按钮和事件监听器
    const importProblemBtn = document.createElement('button');
    importProblemBtn.className = 'btn btn-sm btn-outline-info me-2';
    importProblemBtn.textContent = '导入题目';
    importProblemBtn.id = 'importProblemBtn';
    importProblemBtn.addEventListener('click', showImportProblemModal);
    
    // 添加AI生成题目的按钮和事件监听器
    const generateAIProblemBtn = document.createElement('button');
    generateAIProblemBtn.className = 'btn btn-sm btn-outline-success me-2';
    generateAIProblemBtn.textContent = 'AI生成题目';
    generateAIProblemBtn.id = 'generateAIProblemBtn';
    generateAIProblemBtn.addEventListener('click', showGenerateAIProblemModal);
    
    // 将按钮添加到题目列表卡片的头部
    const cardHeader = document.querySelector('#problems .card-header div');
    if (cardHeader) {
        cardHeader.insertBefore(generateAIProblemBtn, cardHeader.firstChild);
        cardHeader.insertBefore(importProblemBtn, cardHeader.firstChild);
    }
    
    // 检查用户登录状态
    checkUserLoginStatus();
    
    // 加载题目列表
    loadProblems();
});

// 检查用户登录状态
async function checkUserLoginStatus() {
    try {
        // 尝试从localStorage获取用户信息
        const storedUser = localStorage.getItem('currentUser');
        if (storedUser) {
            currentUser = JSON.parse(storedUser);
            console.log('从本地存储加载用户:', currentUser);
            return;
        }
        
        // 如果没有本地存储的用户，尝试创建一个默认用户
        console.log('尝试创建默认用户...');
        const response = await fetch('/api/users', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: 'guest_' + Math.floor(Math.random() * 10000),
                email: 'guest@example.com',
                password: 'guestpassword'
            })
        });
        
        if (response.ok) {
            const userData = await response.json();
            currentUser = {
                id: userData.id,
                username: userData.username
            };
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            console.log('创建了新用户:', currentUser);
        } else {
            console.warn('无法创建用户，使用默认ID');
        }
    } catch (error) {
        console.error('检查用户登录状态失败:', error);
    }
}

// 加载题目列表
async function loadProblems() {
    try {
        if (!problemsList) {
            console.error('无法找到题目列表元素');
            return;
        }
        
        problemsList.innerHTML = '<tr><td colspan="4" class="text-center">加载中...</td></tr>';

        // 添加详细的错误日志
        console.log('正在加载题目...用户ID:', currentUser.id);
        
        // 使用正确的API路径，添加用户ID作为查询参数
        const url = `/api/problems/status?user_id=${currentUser.id}`;
        console.log('请求URL:', url);
        
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            }
        });
        
        console.log('API响应状态:', response.status);
        
        if (!response.ok) {
            // 尝试获取详细错误信息
            let errorText = '';
            try {
                const errorData = await response.json();
                errorText = errorData.error || response.statusText;
            } catch (e) {
                errorText = await response.text() || response.statusText;
            }
            
            console.error('加载题目失败:', errorText);
            
            // 如果是用户ID错误，尝试重新创建用户
            if (errorText.includes('用户ID') || errorText.includes('user')) {
                console.log('尝试重新创建用户并重试...');
                localStorage.removeItem('currentUser');
                await checkUserLoginStatus();
                await loadProblems();
                return;
            }
            
            throw new Error(`加载题目失败 (${response.status}): ${errorText}`);
        }

        const problems = await response.json();
        console.log('成功加载题目数量:', problems.length);
        
        renderProblems(problems);
    } catch (error) {
        console.error('加载题目错误:', error);
        problemsList.innerHTML = `<tr><td colspan="4" class="text-center text-danger">加载题目失败: ${error.message}</td></tr>`;
    }
}

// 渲染题目列表
function renderProblems(problems) {
    if (!problemsList) {
        console.error('无法找到题目列表元素');
        return;
    }
    
    if (!problems || problems.length === 0) {
        problemsList.innerHTML = '<tr><td colspan="4" class="text-center">没有找到题目</td></tr>';
        return;
    }

    problemsList.innerHTML = '';

    problems.forEach(problem => {
        const row = document.createElement('tr');
        
        // 设置状态标记
        let statusClass = 'status-unsolved';
        let statusText = '未尝试';
        
        if (problem.solved) {
            statusClass = 'status-solved';
            statusText = '已解决';
        } else if (problem.attempted) {
            statusClass = 'status-attempted';
            statusText = '尝试中';
        }
        
        row.innerHTML = `
            <td>${problem.id}</td>
            <td>
                <a href="#problem-${problem.id}" class="problem-link" data-problem-id="${problem.id}">
                    ${problem.title}
                </a>
            </td>
            <td>${problem.difficulty}</td>
            <td><span class="status-indicator ${statusClass}"></span> ${statusText}</td>
        `;
        
        // 添加点击事件
        const problemLink = row.querySelector('.problem-link');
        problemLink.addEventListener('click', (e) => {
            e.preventDefault();
            loadProblemDetail(problem.id);
        });
        
        problemsList.appendChild(row);
    });
}

// 加载题目详情
async function loadProblemDetail(problemId) {
    try {
        console.log(`正在加载题目详情 ID=${problemId}, 用户ID=${currentUser.id}`);
        
        const response = await fetch(`/api/problems/${problemId}/status?user_id=${currentUser.id}`, {
            method: 'GET',
            headers: {
                'Accept': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`加载题目详情失败 (${response.status})`);
        }
        
        const problem = await response.json();
        
        // 显示题目详情（这里可以根据需要实现具体的显示逻辑）
        showProblemDetail(problem);
    } catch (error) {
        console.error('加载题目详情失败:', error);
        alert(`加载题目详情失败: ${error.message}`);
    }
}

// 显示题目详情
function showProblemDetail(problem) {
    // 记录问题对象以便调试
    console.log('问题对象:', problem);
    
    // 安全地获取属性，防止undefined错误
    const title = problem.title || '未命名题目';
    const difficulty = problem.difficulty || '未知难度';
    const timeLimit = problem.time_limit || 1000;
    const memoryLimit = problem.memory_limit || 65536;
    const description = problem.description || '无题目描述';
    
    // 创建模态框显示题目详情
    const modal = document.createElement('div');
    modal.className = 'modal fade';
    modal.id = 'problemDetailModal';
    modal.setAttribute('tabindex', '-1');
    modal.setAttribute('aria-labelledby', 'problemDetailModalLabel');
    modal.setAttribute('aria-hidden', 'true');
    
    modal.innerHTML = `
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="problemDetailModalLabel">${title}</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="d-flex justify-content-between mb-3">
                        <span class="badge bg-info">难度: ${difficulty}</span>
                        <span class="badge bg-secondary">时间限制: ${timeLimit}ms</span>
                        <span class="badge bg-secondary">内存限制: ${memoryLimit}KB</span>
                    </div>
                    
                    <h6>题目描述:</h6>
                    <div class="problem-description mb-3">
                        ${description.toString().replace(/\n/g, '<br>')}
                    </div>
                    
                    <h6>样例:</h6>
                    <div class="examples mb-3">
                        ${renderExamples(problem)}
                    </div>
                    
                    <h6>知识点:</h6>
                    <div class="knowledge-tags mb-3">
                        ${renderKnowledgeTags(problem)}
                    </div>
                    
                    <h6>提交代码:</h6>
                    <div class="code-submission">
                        <textarea class="form-control code-editor" id="codeSubmission" rows="10"></textarea>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">关闭</button>
                    <button type="button" class="btn btn-primary" id="submitCodeBtn" data-problem-id="${problem.id}">提交代码</button>
                </div>
            </div>
        </div>
    `;
    
    // 移除现有的模态框（如果有）
    const existingModal = document.getElementById('problemDetailModal');
    if (existingModal) {
        existingModal.remove();
    }
    
    // 添加新的模态框
    document.body.appendChild(modal);
    
    // 初始化Bootstrap模态框
    const problemModal = new bootstrap.Modal(modal);
    problemModal.show();
    
    // 添加提交代码的事件监听器
    const submitCodeBtn = document.getElementById('submitCodeBtn');
    if (submitCodeBtn) {
        submitCodeBtn.addEventListener('click', () => submitCode(problem.id));
    }
}

// 渲染题目样例
function renderExamples(problem) {
    // 检查problem.examples是否存在
    if (!problem.examples || !Array.isArray(problem.examples) || problem.examples.length === 0) {
        // 尝试从test_cases中找到is_example为true的测试用例
        if (problem.test_cases && Array.isArray(problem.test_cases)) {
            const exampleCases = problem.test_cases.filter(tc => tc.is_example);
            if (exampleCases.length > 0) {
                let examplesHtml = '';
                exampleCases.forEach((example, index) => {
                    examplesHtml += `
                        <div class="example mb-3">
                            <h6>样例 ${index + 1}:</h6>
                            <div class="row">
                                <div class="col-md-6">
                                    <h6>输入:</h6>
                                    <pre>${example.input || ''}</pre>
                                </div>
                                <div class="col-md-6">
                                    <h6>输出:</h6>
                                    <pre>${example.output || ''}</pre>
                                </div>
                            </div>
                        </div>
                    `;
                });
                return examplesHtml;
            }
        }
        return '<p>无样例</p>';
    }
    
    let examplesHtml = '';
    
    problem.examples.forEach((example, index) => {
        examplesHtml += `
            <div class="example mb-3">
                <h6>样例 ${index + 1}:</h6>
                <div class="row">
                    <div class="col-md-6">
                        <h6>输入:</h6>
                        <pre>${example.input || ''}</pre>
                    </div>
                    <div class="col-md-6">
                        <h6>输出:</h6>
                        <pre>${example.output || ''}</pre>
                    </div>
                </div>
            </div>
        `;
    });
    
    return examplesHtml;
}

// 渲染知识点标签
function renderKnowledgeTags(problem) {
    if (!problem.knowledge_tag || !Array.isArray(problem.knowledge_tag) || problem.knowledge_tag.length === 0) {
        return '<p>无标签</p>';
    }
    
    return problem.knowledge_tag.map(tag => 
        `<span class="badge bg-primary me-1">${tag}</span>`
    ).join(' ');
}

// 提交代码
async function submitCode(problemId) {
    try {
        const codeEditor = document.getElementById('codeSubmission');
        const code = codeEditor.value.trim();
        
        if (!code) {
            alert('请输入代码');
            return;
        }
        
        console.log(`正在提交代码，题目ID=${problemId}, 用户ID=${currentUser.id}`);
        
        const response = await fetch(`/api/problems/${problemId}/submissions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                user_id: currentUser.id,
                problem_id: problemId,
                language: 'cpp',
                code: code
            })
        });
        
        if (!response.ok) {
            let errorMessage = `提交代码失败 (${response.status})`;
            try {
                const errorData = await response.json();
                errorMessage += ': ' + (errorData.error || '');
            } catch (e) {
                // 如果无法解析JSON，使用默认错误信息
            }
            throw new Error(errorMessage);
        }
        
        const result = await response.json();
        
        // 关闭题目详情模态框
        const problemModal = bootstrap.Modal.getInstance(document.getElementById('problemDetailModal'));
        if (problemModal) {
            problemModal.hide();
        }
        
        // 显示提交结果
        alert(`代码提交成功！提交ID: ${result.id}`);
        
        // 切换到提交记录标签页
        document.getElementById('submissions-tab').click();
        
        // 刷新题目列表以更新状态
        loadProblems();
    } catch (error) {
        console.error('提交代码失败:', error);
        alert(`提交代码失败: ${error.message}`);
    }
}

// 显示创建题目模态框
function showCreateProblemModal() {
    alert('创建题目功能正在开发中...');
    // 这里可以实现创建题目的模态框
}

// 显示导入题目模态框
function showImportProblemModal() {
    // 创建模态框
    const modal = document.createElement('div');
    modal.className = 'modal fade';
    modal.id = 'importProblemModal';
    modal.setAttribute('tabindex', '-1');
    modal.setAttribute('aria-labelledby', 'importProblemModalLabel');
    modal.setAttribute('aria-hidden', 'true');
    
    // 创建一个示例题目的JSON
    const sampleProblemJson = JSON.stringify({
        "problems": [
            {
                "title": "数组求和",
                "description": "给定一个整数数组，求所有元素的和。\n\n**输入格式**\n第一行包含一个整数 n (1 ≤ n ≤ 100)，表示数组的长度。\n第二行包含 n 个整数，表示数组中的元素，每个元素的绝对值不超过 1000。\n\n**输出格式**\n输出一个整数，表示数组所有元素的和。",
                "difficulty": "Easy",
                "time_limit": 1000,
                "memory_limit": 65536,
                "knowledge_tag": ["数组", "基础算法"],
                "test_cases": [
                    {
                        "input": "5\n1 2 3 4 5",
                        "output": "15",
                        "is_example": true
                    },
                    {
                        "input": "3\n-1 2 -3",
                        "output": "-2",
                        "is_example": true
                    },
                    {
                        "input": "1\n0",
                        "output": "0",
                        "is_example": false
                    }
                ]
            }
        ]
    }, null, 2);
    
    modal.innerHTML = `
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="importProblemModalLabel">导入题目</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="mb-3">
                        <label for="importFormat" class="form-label">选择导入格式</label>
                        <select class="form-select" id="importFormat">
                            <option value="json">JSON</option>
                            <option value="xml" disabled>XML (暂不支持)</option>
                        </select>
                    </div>
                    
                    <div class="mb-3">
                        <label for="importData" class="form-label">导入数据</label>
                        <div class="d-flex justify-content-end mb-1">
                            <button class="btn btn-sm btn-outline-secondary" id="insertSampleBtn">插入示例</button>
                        </div>
                        <textarea class="form-control" id="importData" rows="15" placeholder="请输入JSON格式的题目数据"></textarea>
                    </div>
                    
                    <div class="mb-3 form-check">
                        <input type="checkbox" class="form-check-input" id="importValidateCheck" checked>
                        <label class="form-check-label" for="importValidateCheck">导入前验证</label>
                    </div>
                    
                    <div class="alert alert-info">
                        <h6>导入格式说明:</h6>
                        <p>JSON应包含一个"problems"数组，每个题目需包含以下字段:</p>
                        <ul>
                            <li><code>title</code>: 题目标题</li>
                            <li><code>description</code>: 题目描述 (支持基本Markdown格式)</li>
                            <li><code>difficulty</code>: 难度 (Easy, Medium, Hard)</li>
                            <li><code>time_limit</code>: 时间限制 (毫秒)</li>
                            <li><code>memory_limit</code>: 内存限制 (KB)</li>
                            <li><code>knowledge_tag</code>: 知识点标签数组</li>
                            <li><code>test_cases</code>: 测试用例数组，每个测试用例包含:</li>
                            <ul>
                                <li><code>input</code>: 输入数据</li>
                                <li><code>output</code>: 期望输出</li>
                                <li><code>is_example</code>: 是否为展示给用户的样例</li>
                            </ul>
                        </ul>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                    <button type="button" class="btn btn-primary" id="confirmImportBtn">导入</button>
                </div>
            </div>
        </div>
    `;
    
    // 移除现有的模态框（如果有）
    const existingModal = document.getElementById('importProblemModal');
    if (existingModal) {
        existingModal.remove();
    }
    
    // 添加新的模态框
    document.body.appendChild(modal);
    
    // 初始化Bootstrap模态框
    const importModal = new bootstrap.Modal(modal);
    importModal.show();
    
    // 添加插入示例按钮的事件监听器
    const insertSampleBtn = document.getElementById('insertSampleBtn');
    if (insertSampleBtn) {
        insertSampleBtn.addEventListener('click', () => {
            const importData = document.getElementById('importData');
            if (importData) {
                importData.value = sampleProblemJson;
            }
        });
    }
    
    // 添加确认导入的事件监听器
    const confirmImportBtn = document.getElementById('confirmImportBtn');
    if (confirmImportBtn) {
        confirmImportBtn.addEventListener('click', importProblems);
    }
}

// 导入题目
async function importProblems() {
    try {
        const importFormat = document.getElementById('importFormat').value;
        const importData = document.getElementById('importData').value.trim();
        const shouldValidate = document.getElementById('importValidateCheck').checked;
        
        if (!importData) {
            alert('请输入导入数据');
            return;
        }
        
        let parsedData;
        try {
            parsedData = JSON.parse(importData);
        } catch (error) {
            alert(`JSON解析失败: ${error.message}`);
            return;
        }
        
        if (shouldValidate) {
            // 基本验证
            if (!parsedData.problems || !Array.isArray(parsedData.problems) || parsedData.problems.length === 0) {
                alert('无效的导入数据格式，必须包含problems数组');
                return;
            }
            
            // 验证每个题目
            for (const problem of parsedData.problems) {
                if (!problem.title || !problem.description) {
                    alert('题目必须包含标题和描述');
                    return;
                }
                
                if (!problem.test_cases || !Array.isArray(problem.test_cases) || problem.test_cases.length === 0) {
                    alert('每个题目必须至少包含一个测试用例');
                    return;
                }
            }
        }
        
        console.log('正在导入题目...');
        
        const response = await fetch('/api/problems/import', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(parsedData)
        });
        
        if (!response.ok) {
            // 尝试获取详细错误信息
            let errorText = '';
            try {
                const errorData = await response.json();
                errorText = errorData.error || response.statusText;
            } catch (e) {
                errorText = await response.text() || response.statusText;
            }
            
            throw new Error(`导入题目失败 (${response.status}): ${errorText}`);
        }
        
        const result = await response.json();
        
        // 关闭导入模态框
        const importModal = bootstrap.Modal.getInstance(document.getElementById('importProblemModal'));
        if (importModal) {
            importModal.hide();
        }
        
        // 显示导入结果
        alert(`成功导入 ${result.imported_count || result.length || 0} 个题目`);
        
        // 刷新题目列表
        loadProblems();
    } catch (error) {
        console.error('导入题目失败:', error);
        alert(`导入题目失败: ${error.message}`);
    }
}

// 获取URL参数
function getUrlParam(param) {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get(param);
}

// 格式化日期时间
function formatDateTime(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString();
}
