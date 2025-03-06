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

// 显示AI生成题目模态框
function showGenerateAIProblemModal() {
    // 创建模态框HTML
    const modalHTML = `
    <div class="modal fade" id="generateAIProblemModal" tabindex="-1" aria-labelledby="generateAIProblemModalLabel" aria-hidden="true">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="generateAIProblemModalLabel">AI生成题目（基于信息赛大纲）</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="alert alert-info">
                        <i class="bi bi-info-circle"></i> 您只需选择<strong>大纲范围</strong>和<strong>至少一个知识点</strong>，其他字段均可留空，系统将自动生成相应内容。
                    </div>
                    <form id="generateAIProblemForm">
                        <!-- 大纲范围和题目主题 -->
                        <div class="row mb-3">
                            <div class="col-md-6">
                                <label for="outlineSectionSelect" class="form-label">大纲范围</label>
                                <select class="form-select" id="outlineSectionSelect" required>
                                    <option value="" selected disabled>请选择章节...</option>
                                    <!-- 入门级章节 -->
                                    <option value="2.1">2.1 入门级</option>
                                    <option value="2.1.1">2.1.1 基础知识与编程环境</option>
                                    <option value="2.1.2">2.1.2 C++程序设计</option>
                                    <option value="2.1.3">2.1.3 数据结构</option>
                                    <option value="2.1.4">2.1.4 算法</option>
                                    <option value="2.1.5">2.1.5 数学与其他</option>
                                    <!-- 提高级章节 -->
                                    <option value="2.2">2.2 提高级</option>
                                    <option value="2.2.1">2.2.1 基础知识与编程环境</option>
                                    <option value="2.2.2">2.2.2 C++程序设计</option>
                                    <option value="2.2.3">2.2.3 数据结构</option>
                                    <option value="2.2.4">2.2.4 算法</option>
                                    <option value="2.2.5">2.2.5 数学与其他</option>
                                    <!-- NOI级章节 -->
                                    <option value="2.3">2.3 NOI级</option>
                                    <option value="2.3.1">2.3.1 C++程序设计</option>
                                    <option value="2.3.2">2.3.2 数据结构</option>
                                    <option value="2.3.3">2.3.3 算法</option>
                                    <option value="2.3.4">2.3.4 数学与其他</option>
                                </select>
                            </div>
                            <div class="col-md-6">
                                <label for="problemTitleInput" class="form-label">题目主题 (可留空)</label>
                                <input type="text" class="form-control" id="problemTitleInput" placeholder="可留空，AI将自动生成标题...">
                            </div>
                        </div>
                        
                        <!-- 知识点选择区 -->
                        <div class="mb-3">
                            <label class="form-label">知识点选择</label>
                            <div id="knowledgePointsContainer" class="border p-2 mb-2" style="max-height: 150px; overflow-y: auto;">
                                <div class="spinner-border spinner-border-sm" role="status">
                                    <span class="visually-hidden">Loading...</span>
                                </div>
                                <span class="text-muted">请先选择大纲章节...</span>
                            </div>
                        </div>
                        
                        <!-- 题目难度和类型 -->
                        <div class="row mb-3">
                            <div class="col-md-6">
                                <label for="difficultySelect" class="form-label">难度级别</label>
                                <select class="form-select" id="difficultySelect">
                                    <option value="Easy">简单</option>
                                    <option value="Medium" selected>中等</option>
                                    <option value="Hard">困难</option>
                                </select>
                            </div>
                            <div class="col-md-6">
                                <label class="form-label">题目类型</label>
                                <div>
                                    <div class="form-check form-check-inline">
                                        <input class="form-check-input" type="radio" name="problemType" id="algorithmProblem" value="算法题" checked>
                                        <label class="form-check-label" for="algorithmProblem">算法题</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input class="form-check-input" type="radio" name="problemType" id="dataStructureProblem" value="数据结构题">
                                        <label class="form-check-label" for="dataStructureProblem">数据结构题</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input class="form-check-input" type="radio" name="problemType" id="mathProblem" value="数学题">
                                        <label class="form-check-label" for="mathProblem">数学题</label>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <!-- 额外要求 -->
                        <div class="mb-3">
                            <label for="additionalRequirementsInput" class="form-label">额外要求 (可留空)</label>
                            <textarea class="form-control" id="additionalRequirementsInput" rows="2" placeholder="可留空，有特殊要求可在此输入..."></textarea>
                        </div>
                        
                        <!-- 测试用例数和生成选项 -->
                        <div class="row mb-3">
                            <div class="col-md-4">
                                <label for="testCaseCountInput" class="form-label">测试用例数量</label>
                                <input type="number" class="form-control" id="testCaseCountInput" value="3" min="1" max="10">
                            </div>
                            <div class="col-md-8">
                                <label class="form-label">其他选项</label>
                                <div>
                                    <div class="form-check form-check-inline">
                                        <input class="form-check-input" type="checkbox" id="includeReference" value="includeReference" checked>
                                        <label class="form-check-label" for="includeReference">生成参考解答</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input class="form-check-input" type="checkbox" id="includeAnalysis" value="includeAnalysis" checked>
                                        <label class="form-check-label" for="includeAnalysis">生成思维训练分析</label>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- 模型选择 -->
                        <div class="mb-3">
                            <label for="modelTypeSelect" class="form-label">选择模型</label>
                            <select class="form-select" id="modelTypeSelect">
                                <option value="deepseek" selected>DeepSeek Coder (默认)</option>
                                <option value="deepseek_silicon">DeepSeek Coder 硅基流动 (高级)</option>
                                <option value="openai">OpenAI</option>
                            </select>
                            <div class="form-text">硅基流动模型为33B大模型，能生成更加复杂和精确的题目</div>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                    <button type="button" class="btn btn-primary" id="generateAIProblemBtn">生成题目</button>
                </div>
            </div>
        </div>
    </div>
    `;

    // 添加到页面
    document.body.insertAdjacentHTML('beforeend', modalHTML);
    
    // 获取模态框实例
    const modal = document.getElementById('generateAIProblemModal');
    const modalInstance = new bootstrap.Modal(modal);
    
    // 章节选择改变时加载对应知识点
    const sectionSelect = document.getElementById('outlineSectionSelect');
    sectionSelect.addEventListener('change', loadKnowledgePoints);
    
    // 生成按钮事件
    const generateButton = document.getElementById('generateAIProblemBtn');
    generateButton.addEventListener('click', generateAIProblem);
    
    // 模态框关闭事件
    modal.addEventListener('hidden.bs.modal', function() {
        document.getElementById('generateAIProblemModal').remove();
    });
    
    // 显示模态框
    modalInstance.show();
}

// 加载知识点
async function loadKnowledgePoints() {
    const sectionId = document.getElementById('outlineSectionSelect').value;
    const container = document.getElementById('knowledgePointsContainer');
    
    // 显示加载状态
    container.innerHTML = `
        <div class="d-flex align-items-center">
            <div class="spinner-border spinner-border-sm me-2" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
            <span>加载知识点中...</span>
        </div>
    `;
    
    try {
        // 发送请求获取知识点
        const response = await fetch(`/api/outline/knowledge-points?section=${sectionId}`);
        
        if (!response.ok) {
            throw new Error(`获取知识点失败 (${response.status})`);
        }
        
        const data = await response.json();
        console.log('获取到知识点:', data);
        
        // 如果没有知识点，提供默认值
        if (!data || data.length === 0) {
            console.log('没有找到知识点，使用硬编码的默认知识点');
            renderDefaultKnowledgePoints(container, sectionId);
            return;
        }
        
        // 渲染知识点选择框
        let html = '';
        data.forEach(item => {
            const id = `kp_${item.Knowledge.replace(/\s+/g, '_')}`;
            html += `
                <div class="form-check mb-1">
                    <input class="form-check-input" type="checkbox" name="knowledgePoints" id="${id}" value="${item.Knowledge}">
                    <label class="form-check-label" for="${id}">
                        ${item.Knowledge} ${item.Tags && item.Tags.length > 0 ? 
                            `<span class="text-muted small">(${item.Tags.join(', ')})</span>` : ''}
                    </label>
                </div>
            `;
        });
        
        container.innerHTML = html;
    } catch (error) {
        console.error('加载知识点失败:', error);
        console.log('使用硬编码的默认知识点');
        renderDefaultKnowledgePoints(container, sectionId);
    }
}

// 根据章节提供默认知识点
function renderDefaultKnowledgePoints(container, sectionId) {
    // 默认知识点映射，根据大纲章节ID提供相应的默认知识点，基于2024年信息赛大纲.md
    const defaultKnowledgePoints = {
        // 入门级
        "2.1": ["程序设计基础", "数据结构基础", "算法基础", "数学基础"],
        "2.1.1": ["计算机的基本构成", "操作系统基本概念", "计算机网络基本概念", "程序设计语言概念", "开发环境配置", "编译器使用"],
        "2.1.2": ["标识符与关键字", "基本数据类型", "程序基本语句", "基本运算", "数学库函数", "结构化程序设计", "数组", "字符串处理", "函数与递归", "结构体与联合体", "指针类型", "文件读写", "STL模板"],
        "2.1.3": ["链表", "栈", "队列", "树的定义", "二叉树", "完全二叉树", "图的定义", "图的表示与存储"],
        "2.1.4": ["枚举法", "模拟法", "贪心法", "递推法", "递归法", "二分法", "高精度运算", "排序算法", "深度优先搜索", "广度优先搜索", "图论基础算法", "动态规划基础"],
        "2.1.5": ["整数与实数运算", "进制转换", "初等代数", "初等几何", "整除与因数", "取整", "模运算", "素数筛法", "组合数学基础", "ASCII码"],
        // 提高级
        "2.2": ["C++高级特性", "高级数据结构", "高级算法", "高级数学"],
        "2.2.1": ["Linux系统命令", "文本编辑工具", "编译选项", "GDB调试工具"],
        "2.2.2": ["类与面向对象", "STL容器", "算法模板库"],
        "2.2.3": ["双端栈", "单调队列", "优先队列", "并查集", "二叉堆", "树状数组", "线段树", "字典树", "平衡树", "图论扩展", "哈希表"],
        "2.2.4": ["时间复杂度分析", "空间复杂度分析", "分治算法", "字符串匹配", "高级搜索技巧", "高级图论算法", "树上算法", "高级动态规划"],
        "2.2.5": ["高等代数", "高等几何", "同余式", "欧拉定理", "费马小定理", "组合数学进阶", "线性代数基础", "高斯消元"],
        // NOI级
        "2.3": ["高级程序设计", "复杂数据结构", "竞赛级算法", "高级数学"],
        "2.3.1": ["面向对象编程思想", "设计模式", "高级C++特性"],
        "2.3.2": ["块状链表", "跳跃表", "树链剖分", "动态树", "二维线段树", "树套树", "可持久化数据结构"],
        "2.3.3": ["分块算法", "离线处理", "复杂分治", "平衡规划", "高级字符串算法", "网络流", "复杂动态规划"],
        "2.3.4": ["原根和指数", "Dirichlet卷积", "莫比乌斯反演", "多项式微积分", "傅里叶变换", "概率论", "博弈论", "计算几何"]
    };
    
    // 如果没有指定章节或该章节未定义默认知识点，使用入门级知识点
    const knowledgePoints = defaultKnowledgePoints[sectionId] || defaultKnowledgePoints["2.1"];
    
    let html = '';
    knowledgePoints.forEach(knowledge => {
        const id = `kp_${knowledge.replace(/\s+/g, '_')}`;
        html += `
            <div class="form-check mb-1">
                <input class="form-check-input" type="checkbox" name="knowledgePoints" id="${id}" value="${knowledge}">
                <label class="form-check-label" for="${id}">
                    ${knowledge}
                </label>
            </div>
        `;
    });
    
    container.innerHTML = html;
}

async function generateAIProblem() {
    // 获取表单元素
    const form = document.getElementById('generateAIProblemForm');
    const outlineSection = document.getElementById('outlineSectionSelect').value;
    const problemTitle = document.getElementById('problemTitleInput').value || ""; // 允许为空
    const difficulty = document.getElementById('difficultySelect').value;
    const testCaseCount = parseInt(document.getElementById('testCaseCountInput').value);
    const includeReference = document.getElementById('includeReference').checked;
    const includeAnalysis = document.getElementById('includeAnalysis').checked;
    const additionalRequirements = document.getElementById('additionalRequirementsInput').value || ""; // 允许为空
    const modelType = document.getElementById('modelTypeSelect').value;
    
    // 获取题目类型
    const problemTypeElements = document.querySelectorAll('input[name="problemType"]');
    let problemType = '';
    for (const element of problemTypeElements) {
        if (element.checked) {
            problemType = element.value;
            break;
        }
    }
    
    // 获取选中的知识点
    const knowledgePointsElements = document.querySelectorAll('input[name="knowledgePoints"]:checked');
    const knowledgePoints = Array.from(knowledgePointsElements).map(el => el.value);
    
    // 基本验证
    if (!outlineSection) {
        alert('请选择大纲范围');
        return;
    }
    
    if (knowledgePoints.length === 0) {
        alert('请至少选择一个知识点');
        return;
    }
    
    // 显示加载状态
    const generateButton = document.getElementById('generateAIProblemBtn');
    const originalButtonText = generateButton.textContent;
    generateButton.disabled = true;
    generateButton.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> 生成中...';
    
    try {
        // 准备请求数据
        const requestData = {
            title: problemTitle, // 如果为空，后端会自动生成
            outline_section: outlineSection,
            knowledge_points: knowledgePoints,
            difficulty: difficulty,
            problem_type: problemType,
            additional_reqs: additionalRequirements, // 如果为空，后端不会有额外要求
            test_case_count: testCaseCount,
            include_reference_solution: includeReference,
            include_analysis: includeAnalysis,
            model_type: modelType,
            auto_generate: problemTitle === "" // 新增字段，表示是否由AI自动生成标题和内容
        };
        
        console.log('生成题目请求数据:', requestData);
        
        // 向后端API发送请求
        const response = await fetch('/api/problems/generate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify(requestData)
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error('生成题目API响应错误:', errorText);
            let errorMessage = '生成题目失败';
            try {
                const errorData = JSON.parse(errorText);
                errorMessage = errorData.error || errorData.message || errorMessage;
            } catch (e) {
                // 如果不是JSON格式，使用原始错误文本
                errorMessage = errorText || errorMessage;
            }
            throw new Error(errorMessage);
        }
        
        const data = await response.json();
        console.log('生成题目响应数据:', data);
        
        // 关闭生成面板
        bootstrap.Modal.getInstance(document.getElementById('generateAIProblemModal')).hide();
        
        // 显示生成结果
        showGeneratedProblemResult(data);
        
    } catch (error) {
        console.error('生成题目出错:', error);
        alert(`生成题目失败: ${error.message}`);
    } finally {
        // 恢复按钮状态
        generateButton.disabled = false;
        generateButton.textContent = originalButtonText;
    }
}

function showGeneratedProblemResult(problem) {
    // 创建展示生成结果的模态框
    const modalHTML = `
    <div class="modal fade" id="generatedProblemModal" tabindex="-1" aria-labelledby="generatedProblemModalLabel" aria-hidden="true">
        <div class="modal-dialog modal-xl">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="generatedProblemModalLabel">生成的题目</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="card mb-4">
                        <div class="card-header d-flex justify-content-between align-items-center">
                            <h4>${problem.title || '未命名题目'}</h4>
                            <div>
                                <span class="badge bg-${problem.difficulty === 'Easy' ? 'success' : problem.difficulty === 'Medium' ? 'warning' : 'danger'}">${problem.difficulty}</span>
                                <span class="ms-2">时间限制: ${problem.time_limit || 1000}ms</span>
                                <span class="ms-2">内存限制: ${problem.memory_limit || 256000}KB</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h5>题目描述</h5>
                            <div class="mb-4">${problem.description ? problem.description.replace(/\n/g, '<br>') : '无描述'}</div>
                            
                            <h5>输入格式</h5>
                            <div class="mb-4">${problem.input_format ? problem.input_format.replace(/\n/g, '<br>') : '无输入格式说明'}</div>
                            
                            <h5>输出格式</h5>
                            <div class="mb-4">${problem.output_format ? problem.output_format.replace(/\n/g, '<br>') : '无输出格式说明'}</div>
                            
                            <h5>样例</h5>
                            ${renderGeneratedExamples(problem)}
                            
                            <h5>数据范围与提示</h5>
                            <div class="mb-4">${problem.constraints ? problem.constraints.replace(/\n/g, '<br>') : '无数据范围说明'}</div>
                            
                            ${problem.knowledge_tag && problem.knowledge_tag.length > 0 ? 
                                `<div class="mt-3">
                                    <h5>相关知识点</h5>
                                    <div>${problem.knowledge_tag.map(tag => `<span class="badge bg-info me-2">${tag}</span>`).join('')}</div>
                                </div>` : ''
                            }
                            
                            ${problem.reference_solution ? 
                                `<div class="mt-4">
                                    <h5 class="text-success">参考解答</h5>
                                    <div class="mb-3">${problem.reference_solution.replace(/\n/g, '<br>')}</div>
                                </div>` : ''
                            }
                            
                            ${problem.thinking_analysis ? 
                                `<div class="mt-4">
                                    <h5 class="text-primary">思维训练分析</h5>
                                    <div class="mb-3">${problem.thinking_analysis.replace(/\n/g, '<br>')}</div>
                                </div>` : ''
                            }
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">关闭</button>
                    <button type="button" class="btn btn-warning" id="editGeneratedProblemBtn">编辑后保存</button>
                    <button type="button" class="btn btn-success" id="saveGeneratedProblemBtn">直接保存</button>
                </div>
            </div>
        </div>
    </div>
    `;

    // 添加到页面
    document.body.insertAdjacentHTML('beforeend', modalHTML);

    // 获取模态框实例
    const modal = document.getElementById('generatedProblemModal');
    const modalInstance = new bootstrap.Modal(modal);

    // 绑定按钮事件
    document.getElementById('saveGeneratedProblemBtn').addEventListener('click', () => {
        saveGeneratedProblem(problem);
    });

    document.getElementById('editGeneratedProblemBtn').addEventListener('click', () => {
        modalInstance.hide();
        editGeneratedProblem(problem);
    });

    // 模态框关闭时移除
    modal.addEventListener('hidden.bs.modal', function() {
        document.getElementById('generatedProblemModal').remove();
    });

    // 显示模态框
    modalInstance.show();
}

// 渲染生成的样例
function renderGeneratedExamples(problem) {
    if (!problem.test_cases || problem.test_cases.length === 0) {
        return '<div class="text-muted">无样例</div>';
    }

    let html = '';
    const exampleCases = problem.test_cases.filter(tc => tc.is_example);
    
    if (exampleCases.length === 0) {
        // 如果没有标记为样例的测试用例，就使用前两个
        const casesToShow = problem.test_cases.slice(0, 2);
        
        casesToShow.forEach((testCase, index) => {
            html += `
            <div class="mb-3">
                <div class="card">
                    <div class="card-header">样例 ${index + 1}</div>
                    <div class="card-body">
                        <div class="row">
                            <div class="col-md-6">
                                <h6>输入</h6>
                                <pre class="border p-2 bg-light">${testCase.input || '无'}</pre>
                            </div>
                            <div class="col-md-6">
                                <h6>输出</h6>
                                <pre class="border p-2 bg-light">${testCase.output || '无'}</pre>
                            </div>
                        </div>
                    </div>
                </div>
            </div>`;
        });
    } else {
        // 使用标记为样例的测试用例
        exampleCases.forEach((testCase, index) => {
            html += `
            <div class="mb-3">
                <div class="card">
                    <div class="card-header">样例 ${index + 1}</div>
                    <div class="card-body">
                        <div class="row">
                            <div class="col-md-6">
                                <h6>输入</h6>
                                <pre class="border p-2 bg-light">${testCase.input || '无'}</pre>
                            </div>
                            <div class="col-md-6">
                                <h6>输出</h6>
                                <pre class="border p-2 bg-light">${testCase.output || '无'}</pre>
                            </div>
                        </div>
                    </div>
                </div>
            </div>`;
        });
    }
    
    return html;
}

// 保存生成的题目
async function saveGeneratedProblem(problem) {
    try {
        // 准备保存的数据
        const requestData = {
            user_id: currentUser.id,
            problem: {
                title: problem.title,
                description: problem.description,
                difficulty: problem.difficulty,
                time_limit: problem.time_limit,
                memory_limit: problem.memory_limit,
                knowledge_tag: problem.knowledge_tag,
                test_cases: problem.test_cases.map(tc => ({
                    input: tc.input,
                    output: tc.output,
                    is_example: tc.is_example !== undefined ? tc.is_example : true
                }))
            }
        };
        
        console.log('保存生成的题目:', requestData);
        
        // 发送请求到后端API
        const response = await fetch('/api/problems', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestData)
        });
        
        if (!response.ok) {
            throw new Error(`保存题目失败 (${response.status})`);
        }
        
        const result = await response.json();
        
        // 关闭结果模态框
        const resultModal = bootstrap.Modal.getInstance(document.getElementById('generatedProblemModal'));
        if (resultModal) {
            resultModal.hide();
        }
        
        alert(`题目"${problem.title}"已成功保存！ID: ${result.id}`);
        
        // 刷新题目列表
        loadProblems();
        
    } catch (error) {
        console.error('保存生成的题目失败:', error);
        alert(`保存题目失败: ${error.message}`);
    }
}

// 编辑生成的题目
function editGeneratedProblem(problem) {
    // 创建编辑模态框
    const modal = document.createElement('div');
    modal.className = 'modal fade';
    modal.id = 'editGeneratedProblemModal';
    modal.setAttribute('tabindex', '-1');
    modal.setAttribute('aria-labelledby', 'editGeneratedProblemModalLabel');
    modal.setAttribute('aria-hidden', 'true');
    
    // 测试用例HTML
    let testCasesHtml = '';
    if (problem.test_cases && problem.test_cases.length > 0) {
        problem.test_cases.forEach((testCase, index) => {
            testCasesHtml += `
                <div class="test-case-item mb-3">
                    <h6>测试用例 ${index + 1}:</h6>
                    <div class="row">
                        <div class="col-md-6">
                            <label for="testCaseInput${index}" class="form-label">输入:</label>
                            <textarea class="form-control" id="testCaseInput${index}" rows="3">${testCase.input || ''}</textarea>
                        </div>
                        <div class="col-md-6">
                            <label for="testCaseOutput${index}" class="form-label">输出:</label>
                            <textarea class="form-control" id="testCaseOutput${index}" rows="3">${testCase.output || ''}</textarea>
                        </div>
                    </div>
                    <div class="form-check mt-2">
                        <input class="form-check-input" type="checkbox" id="isExample${index}" ${testCase.is_example ? 'checked' : ''}>
                        <label class="form-check-label" for="isExample${index}">
                            显示为样例
                        </label>
                    </div>
                </div>
            `;
        });
    }
    
    // 知识点标签选项
    const allKnowledgeTags = [
        "数组", "链表", "栈", "队列", "哈希表", "堆", "二叉树", "图", 
        "排序", "二分查找", "双指针", "滑动窗口", "贪心算法", "分治算法", 
        "动态规划", "回溯算法", "深度优先搜索", "广度优先搜索", "位运算", 
        "数学", "字符串", "模拟"
    ];
    
    const tagCheckboxes = allKnowledgeTags.map(tag => {
        const isChecked = problem.knowledge_tag.includes(tag);
        return `
            <div class="form-check form-check-inline">
                <input class="form-check-input" type="checkbox" id="edit_tag_${tag}" name="edit_knowledge_tags" value="${tag}" ${isChecked ? 'checked' : ''}>
                <label class="form-check-label" for="edit_tag_${tag}">${tag}</label>
            </div>
        `;
    }).join('');
    
    modal.innerHTML = `
        <div class="modal-dialog modal-xl">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="editGeneratedProblemModalLabel">编辑生成的题目</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <form id="editProblemForm">
                        <div class="mb-3">
                            <label for="editTitle" class="form-label">标题</label>
                            <input type="text" class="form-control" id="editTitle" value="${problem.title}">
                        </div>
                        
                        <div class="row mb-3">
                            <div class="col-md-4">
                                <label for="editDifficulty" class="form-label">难度</label>
                                <select class="form-select" id="editDifficulty">
                                    <option value="Easy" ${problem.difficulty === 'Easy' ? 'selected' : ''}>简单 (Easy)</option>
                                    <option value="Medium" ${problem.difficulty === 'Medium' ? 'selected' : ''}>中等 (Medium)</option>
                                    <option value="Hard" ${problem.difficulty === 'Hard' ? 'selected' : ''}>困难 (Hard)</option>
                                </select>
                            </div>
                            <div class="col-md-4">
                                <label for="editTimeLimit" class="form-label">时间限制 (ms)</label>
                                <input type="number" class="form-control" id="editTimeLimit" value="${problem.time_limit}">
                            </div>
                            <div class="col-md-4">
                                <label for="editMemoryLimit" class="form-label">内存限制 (KB)</label>
                                <input type="number" class="form-control" id="editMemoryLimit" value="${problem.memory_limit}">
                            </div>
                        </div>
                        
                        <div class="mb-3">
                            <label for="editDescription" class="form-label">题目描述</label>
                            <textarea class="form-control" id="editDescription" rows="10">${problem.description}</textarea>
                        </div>
                        
                        <div class="mb-3">
                            <label class="form-label">知识点标签</label>
                            <div class="border p-2 rounded" style="max-height: 150px; overflow-y: auto;">
                                ${tagCheckboxes}
                            </div>
                        </div>
                        
                        <div class="mb-3">
                            <label class="form-label">测试用例</label>
                            <div id="testCasesContainer">
                                ${testCasesHtml}
                            </div>
                            <button type="button" class="btn btn-sm btn-outline-secondary mt-2" id="addTestCaseBtn">
                                <i class="bi bi-plus"></i> 添加测试用例
                            </button>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                    <button type="button" class="btn btn-primary" id="saveEditedProblemBtn">保存</button>
                </div>
            </div>
        </div>
    `;
    
    // 移除现有的模态框（如果有）
    const existingModal = document.getElementById('editGeneratedProblemModal');
    if (existingModal) {
        existingModal.remove();
    }
    
    // 添加新的模态框
    document.body.appendChild(modal);
    
    // 关闭结果模态框
    const resultModal = bootstrap.Modal.getInstance(document.getElementById('generatedProblemModal'));
    if (resultModal) {
        resultModal.hide();
    }
    
    // 初始化编辑模态框
    const editModal = new bootstrap.Modal(modal);
    editModal.show();
    
    // 添加保存编辑的事件监听器
    const saveEditedProblemBtn = document.getElementById('saveEditedProblemBtn');
    if (saveEditedProblemBtn) {
        saveEditedProblem(problem.test_cases.length);
    }
    
    // 添加新测试用例的事件监听器
    const addTestCaseBtn = document.getElementById('addTestCaseBtn');
    if (addTestCaseBtn) {
        addTestCase(problem.test_cases.length);
    }
}

// 添加测试用例
function addTestCase(currentIndex) {
    const testCasesContainer = document.getElementById('testCasesContainer');
    if (!testCasesContainer) return;
    
    const newIndex = currentIndex;
    const newTestCaseDiv = document.createElement('div');
    newTestCaseDiv.className = 'test-case-item mb-3';
    newTestCaseDiv.innerHTML = `
        <h6>测试用例 ${newIndex + 1}:</h6>
        <div class="row">
            <div class="col-md-6">
                <label for="testCaseInput${newIndex}" class="form-label">输入:</label>
                <textarea class="form-control" id="testCaseInput${newIndex}" rows="3"></textarea>
            </div>
            <div class="col-md-6">
                <label for="testCaseOutput${newIndex}" class="form-label">输出:</label>
                <textarea class="form-control" id="testCaseOutput${newIndex}" rows="3"></textarea>
            </div>
        </div>
        <div class="form-check mt-2">
            <input class="form-check-input" type="checkbox" id="isExample${newIndex}" checked>
            <label class="form-check-label" for="isExample${newIndex}">
                显示为样例
            </label>
        </div>
    `;
    
    testCasesContainer.appendChild(newTestCaseDiv);
}

// 保存编辑后的题目
async function saveEditedProblem(originalTestCaseCount) {
    try {
        // 获取编辑后的数据
        const title = document.getElementById('editTitle').value.trim();
        const difficulty = document.getElementById('editDifficulty').value;
        const timeLimit = parseInt(document.getElementById('editTimeLimit').value);
        const memoryLimit = parseInt(document.getElementById('editMemoryLimit').value);
        const description = document.getElementById('editDescription').value.trim();
        
        // 获取选中的知识点标签
        const selectedTags = [];
        document.querySelectorAll('input[name="edit_knowledge_tags"]:checked').forEach(checkbox => {
            selectedTags.push(checkbox.value);
        });
        
        // 表单验证
        if (!title) {
            alert('请输入题目标题');
            return;
        }
        
        if (!description) {
            alert('请输入题目描述');
            return;
        }
        
        if (selectedTags.length === 0) {
            alert('请至少选择一个知识点标签');
            return;
        }
        
        // 获取测试用例
        const testCases = [];
        let currentIndex = 0;
        
        // 先检查原有的测试用例
        for (let i = 0; i < originalTestCaseCount; i++) {
            const inputEl = document.getElementById(`testCaseInput${i}`);
            const outputEl = document.getElementById(`testCaseOutput${i}`);
            const isExampleEl = document.getElementById(`isExample${i}`);
            
            if (inputEl && outputEl) {
                const input = inputEl.value.trim();
                const output = outputEl.value.trim();
                const isExample = isExampleEl ? isExampleEl.checked : true;
                
                if (input && output) {
                    testCases.push({
                        input: input,
                        output: output,
                        is_example: isExample
                    });
                    currentIndex++;
                }
            }
        }
        
        // 然后检查新添加的测试用例
        let newIndex = originalTestCaseCount;
        while (true) {
            const inputEl = document.getElementById(`testCaseInput${newIndex}`);
            const outputEl = document.getElementById(`testCaseOutput${newIndex}`);
            const isExampleEl = document.getElementById(`isExample${newIndex}`);
            
            if (!inputEl || !outputEl) break;
            
            const input = inputEl.value.trim();
            const output = outputEl.value.trim();
            const isExample = isExampleEl ? isExampleEl.checked : true;
            
            if (input && output) {
                testCases.push({
                    input: input,
                    output: output,
                    is_example: isExample
                });
            }
            
            newIndex++;
        }
        
        if (testCases.length === 0) {
            alert('请至少提供一个有效的测试用例');
            return;
        }
        
        // 准备保存的数据
        const requestData = {
            user_id: currentUser.id,
            problem: {
                title: title,
                description: description,
                difficulty: difficulty,
                time_limit: timeLimit,
                memory_limit: memoryLimit,
                knowledge_tag: selectedTags,
                test_cases: testCases
            }
        };
        
        console.log('保存编辑后的题目:', requestData);
        
        // 发送请求到后端API
        const response = await fetch('/api/problems', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestData)
        });
        
        if (!response.ok) {
            throw new Error(`保存题目失败 (${response.status})`);
        }
        
        const result = await response.json();
        
        // 关闭编辑模态框
        const editModal = bootstrap.Modal.getInstance(document.getElementById('editGeneratedProblemModal'));
        if (editModal) {
            editModal.hide();
        }
        
        alert(`题目"${title}"已成功保存！ID: ${result.id}`);
        
        // 刷新题目列表
        loadProblems();
        
    } catch (error) {
        console.error('保存编辑后的题目失败:', error);
        alert(`保存题目失败: ${error.message}`);
    }
} 