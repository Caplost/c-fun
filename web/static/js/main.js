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
        
        // 检查是否为AI生成的题目（通过检查knowledge_tag是否存在且不为空）
        const isAIGenerated = problem.knowledge_tag && problem.knowledge_tag.length > 0;
        
        // 基本的题目信息HTML
        let problemHTML = `
            <td>${problem.id}</td>
            <td>
                <a href="#problem-${problem.id}" class="problem-link" data-problem-id="${problem.id}">
                    ${problem.title}
                </a>`;
                
        // 如果是AI生成的题目，添加知识点标签
        if (isAIGenerated) {
            problemHTML += `
                <div class="mt-1 small problem-knowledge-tags">
                    <strong>知识点:</strong> 
                    ${problem.knowledge_tag.map(tag => 
                        `<span class="badge bg-secondary me-1">${tag}</span>`
                    ).join(' ')}
                </div>`;
        }
        
        problemHTML += `
            </td>
            <td>${problem.difficulty}</td>
            <td><span class="status-indicator ${statusClass}"></span> ${statusText}</td>
        `;
        
        row.innerHTML = problemHTML;
        
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
        
        const data = await response.json();
        
        // 记录完整的响应对象，用于调试
        console.log('API响应数据:', data);
        
        // 显示题目详情（这里应该传递data，而不仅仅是problem）
        showProblemDetail(data);
    } catch (error) {
        console.error('加载题目详情失败:', error);
        alert(`加载题目详情失败: ${error.message}`);
    }
}

// 显示题目详情
function showProblemDetail(data) {
    // 记录问题对象以便调试
    console.log('完整数据对象:', data);
    
    // 从响应中获取problem对象和examples数组
    const problem = data.problem || {};
    const examples = data.examples || [];
    
    // 调试问题详情
    logProblemDetails(problem, examples);
    
    // 安全地获取属性，防止undefined错误
    const title = problem.title || '未命名题目';
    const difficulty = problem.difficulty || '未知难度';
    const timeLimit = problem.time_limit || 1000;
    const memoryLimit = problem.memory_limit || 65536;
    const description = problem.description || '无题目描述';
    
    // 检查是否有参考解答和思维分析
    const referenceSolution = problem.reference_solution || '';
    const thinkingAnalysis = problem.thinking_analysis || '';
    
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
                        ${renderExamples(examples)}
                    </div>
                    
                    <h6>知识点:</h6>
                    <div class="knowledge-tags mb-3">
                        ${renderKnowledgeTags(problem)}
                    </div>
                    
                    ${referenceSolution ? `
                    <div class="mb-3">
                        <button class="btn btn-outline-primary w-100 text-start" type="button" 
                                data-bs-toggle="collapse" data-bs-target="#collapseReferenceSolution" 
                                aria-expanded="false" aria-controls="collapseReferenceSolution">
                            <i class="bi bi-code-square"></i> 参考解答 <span class="small text-muted">(点击展开)</span>
                        </button>
                        <div class="collapse mt-2" id="collapseReferenceSolution">
                            <div class="card card-body">
                                <pre>${referenceSolution.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</pre>
                            </div>
                        </div>
                    </div>
                    ` : ''}
                    
                    ${thinkingAnalysis ? `
                    <div class="mb-3">
                        <button class="btn btn-outline-info w-100 text-start" type="button" 
                                data-bs-toggle="collapse" data-bs-target="#collapseThinkingAnalysis" 
                                aria-expanded="false" aria-controls="collapseThinkingAnalysis">
                            <i class="bi bi-lightbulb"></i> 思维分析 <span class="small text-muted">(点击展开)</span>
                        </button>
                        <div class="collapse mt-2" id="collapseThinkingAnalysis">
                            <div class="card card-body">
                                ${thinkingAnalysis.replace(/\n/g, '<br>')}
                            </div>
                        </div>
                    </div>
                    ` : ''}
                    
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
function renderExamples(examples) {
    // 检查examples是否存在且非空
    if (!examples || !Array.isArray(examples) || examples.length === 0) {
        return '<p>无样例</p>';
    }
    
    let examplesHtml = '';
    
    examples.forEach((example, index) => {
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

// 显示AI生成题目的模态框
function showGenerateAIProblemModal() {
    const modal = document.createElement('div');
    modal.className = 'modal fade';
    modal.id = 'generateAIProblemModal';
    modal.tabIndex = '-1';
    modal.setAttribute('aria-labelledby', 'generateAIProblemModalLabel');
    modal.setAttribute('aria-hidden', 'true');
    
    modal.innerHTML = `
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="generateAIProblemModalLabel">AI生成题目</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <form id="generateAIProblemForm">
                        <div class="mb-3">
                            <label for="aiModelSelect" class="form-label">选择AI模型</label>
                            <select class="form-select" id="aiModelSelect">
                                <option value="deepseek" selected>DeepSeek R1</option>
                                <option value="deepseek_silicon">DeepSeek Silicon</option>
                                <option value="openai">OpenAI (GPT-4o)</option>
                            </select>
                        </div>
                        
                        <div class="mb-3">
                            <label for="generateTitle" class="form-label">题目主题/名称 (可选)</label>
                            <input type="text" class="form-control" id="generateTitle" placeholder="例如：数组排序、二叉树遍历...">
                        </div>
                        
                        <div class="mb-3">
                            <label for="outlineSection" class="form-label">大纲章节</label>
                            <select class="form-select" id="outlineSection">
                                <optgroup label="入门级">
                                    <option value="2.1.1">2.1.1 基础知识与编程环境</option>
                                    <option value="2.1.2">2.1.2 C++程序设计</option>
                                    <option value="2.1.3">2.1.3 数据结构</option>
                                    <option value="2.1.4">2.1.4 算法</option>
                                    <option value="2.1.5">2.1.5 数学与其他</option>
                                </optgroup>
                                <optgroup label="提高级">
                                    <option value="2.2.1">2.2.1 基础知识与编程环境</option>
                                    <option value="2.2.2">2.2.2 C++程序设计</option>
                                    <option value="2.2.3">2.2.3 数据结构</option>
                                    <option value="2.2.4">2.2.4 算法</option>
                                    <option value="2.2.5">2.2.5 数学与其他</option>
                                </optgroup>
                                <optgroup label="NOI级">
                                    <option value="2.3.1">2.3.1 C++程序设计</option>
                                    <option value="2.3.2">2.3.2 数据结构</option>
                                    <option value="2.3.3">2.3.3 算法</option>
                                    <option value="2.3.4">2.3.4 数学与其他</option>
                                </optgroup>
                            </select>
                        </div>
                        
                        <div class="mb-3">
                            <label class="form-label">知识点</label>
                            <div class="mb-2">
                                <button type="button" class="btn btn-sm btn-outline-secondary" id="refreshKnowledgePoints">刷新知识点</button>
                            </div>
                            <div id="knowledgePointsContainer" class="border p-2" style="max-height: 200px; overflow-y: auto;">
                                <div class="d-flex justify-content-center">
                                    <div class="spinner-border spinner-border-sm" role="status">
                                        <span class="visually-hidden">Loading...</span>
                                    </div>
                                    <span class="ms-2">加载知识点中...</span>
                                </div>
                            </div>
                        </div>
                        
                        <div class="mb-3">
                            <label class="form-label">难度</label>
                            <select class="form-select" id="generateDifficulty">
                                <option value="Easy">简单</option>
                                <option value="Medium" selected>中等</option>
                                <option value="Hard">困难</option>
                            </select>
                        </div>
                        
                        <div class="mb-3">
                            <label class="form-label">题目类型</label>
                            <div class="form-check">
                                <input class="form-check-input" type="radio" name="problemType" id="typeAlgorithm" value="Algorithm" checked>
                                <label class="form-check-label" for="typeAlgorithm">算法题</label>
                            </div>
                            <div class="form-check">
                                <input class="form-check-input" type="radio" name="problemType" id="typePractical" value="Practical">
                                <label class="form-check-label" for="typePractical">实际应用题</label>
                            </div>
                        </div>
                        
                        <div class="row mb-3">
                            <div class="col-md-6">
                                <label for="timeComplexity" class="form-label">时间复杂度要求</label>
                                <input type="text" class="form-control" id="timeComplexity" placeholder="例如：O(n)、O(n log n)...">
                            </div>
                            <div class="col-md-6">
                                <label for="spaceComplexity" class="form-label">空间复杂度要求</label>
                                <input type="text" class="form-control" id="spaceComplexity" placeholder="例如：O(n)、O(1)...">
                            </div>
                        </div>
                        
                        <div class="mb-3">
                            <label for="additionalReqs" class="form-label">附加要求 (可选)</label>
                            <textarea class="form-control" id="additionalReqs" rows="3" placeholder="例如：要求使用某种特定算法、限制不能使用某些库函数、输入规模范围..."></textarea>
                        </div>
                        
                        <div class="mb-3">
                            <label for="testCaseCount" class="form-label">测试用例数量</label>
                            <input type="number" class="form-control" id="testCaseCount" min="1" max="20" value="5">
                        </div>
                        
                        <div class="mb-3 form-check">
                            <input type="checkbox" class="form-check-input" id="includeReferenceSolution" checked disabled>
                            <label class="form-check-label" for="includeReferenceSolution">生成参考解答</label>
                        </div>
                        
                        <div class="mb-3 form-check">
                            <input type="checkbox" class="form-check-input" id="includeAnalysis" checked disabled>
                            <label class="form-check-label" for="includeAnalysis">生成思维分析</label>
                        </div>
                    </form>
                    
                    <div id="generateProgress" style="display: none;">
                        <div class="d-flex justify-content-center align-items-center flex-column">
                            <div class="spinner-border text-primary" role="status">
                                <span class="visually-hidden">生成中...</span>
                            </div>
                            <p class="mt-2">AI正在生成题目，请稍候...</p>
                            <p class="text-muted small">根据题目复杂度，可能需要10-60秒</p>
                        </div>
                    </div>
                    
                    <div id="generatedProblemResult" style="display: none;">
                        <div class="alert alert-success">
                            <h5 class="alert-heading">题目生成成功！</h5>
                            <p id="generatedProblemTitle"></p>
                        </div>
                        
                        <div class="card mb-3">
                            <div class="card-header">题目预览</div>
                            <div class="card-body">
                                <h5 id="previewTitle" class="card-title"></h5>
                                <div class="d-flex justify-content-between">
                                    <p class="card-text"><strong>难度:</strong> <span id="previewDifficulty"></span></p>
                                    <p class="card-text" id="previewLimits"></p>
                                </div>
                                <hr>
                                <h6>题目描述:</h6>
                                <div id="previewDescription" class="mb-3"></div>
                                
                                <h6>示例:</h6>
                                <div id="previewExamples" class="mb-3"></div>
                                
                                <h6>知识点标签:</h6>
                                <div id="previewTags" class="mb-3"></div>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="modal-footer" id="generateModalFooter">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">关闭</button>
                    <button type="button" class="btn btn-outline-secondary" id="regenerateProblemBtn" style="display: none;">重新生成</button>
                    <button type="button" class="btn btn-primary" id="startGenerateBtn">开始生成</button>
                    <button type="button" class="btn btn-success" id="saveGeneratedProblemBtn" style="display: none;" disabled>保存到题库</button>
                </div>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    
    // 初始化模态框
    const modalInstance = new bootstrap.Modal(modal);
    modalInstance.show();
    
    // 绑定事件处理
    const refreshKnowledgePointsBtn = document.getElementById('refreshKnowledgePoints');
    if (refreshKnowledgePointsBtn) {
        refreshKnowledgePointsBtn.addEventListener('click', () => {
            loadKnowledgePoints();
        });
    }
    
    // 添加大纲章节变更事件监听
    const outlineSection = document.getElementById('outlineSection');
    if (outlineSection) {
        // 移除已有的事件监听器，防止重复监听
        outlineSection.removeEventListener('change', loadKnowledgePoints);
        // 添加新的事件监听器
        outlineSection.addEventListener('change', loadKnowledgePoints);
        // 初始触发一次加载知识点
        console.log("初始加载知识点...");
        loadKnowledgePoints();
    } else {
        console.error("未找到大纲章节选择器元素!");
    }
    
    // 添加开始生成按钮事件监听器
    const startGenerateBtn = document.getElementById('startGenerateBtn');
    if (startGenerateBtn) {
        startGenerateBtn.addEventListener('click', generateAIProblem);
    }
    
    // 添加保存生成题目按钮的事件监听器
    const saveGeneratedProblemBtn = document.getElementById('saveGeneratedProblemBtn');
    if (saveGeneratedProblemBtn) {
        saveGeneratedProblemBtn.addEventListener('click', saveGeneratedProblem);
    }
    
    // 添加重新生成按钮的事件监听器
    const regenerateProblemBtn = document.getElementById('regenerateProblemBtn');
    if (regenerateProblemBtn) {
        regenerateProblemBtn.addEventListener('click', () => {
            document.getElementById('generateProgress').style.display = 'none';
            document.getElementById('generatedProblemResult').style.display = 'none';
            document.getElementById('generateAIProblemForm').style.display = 'block';
            document.getElementById('generateModalFooter').style.display = 'flex';
        });
    }
}

// 加载知识点
async function loadKnowledgePoints() {
    const outlineSection = document.getElementById('outlineSection');
    const knowledgePointsContainer = document.getElementById('knowledgePointsContainer');
    
    if (!outlineSection || !knowledgePointsContainer) {
        console.error('找不到大纲章节或知识点容器元素');
        return;
    }
    
    const sectionId = outlineSection.value;
    // 添加调试输出
    console.log(`正在为章节 [${sectionId}] 加载知识点`);
    
    // 显示加载中状态
    knowledgePointsContainer.innerHTML = `
        <div class="d-flex justify-content-center">
            <div class="spinner-border spinner-border-sm" role="status">
                <span class="visually-hidden">加载中...</span>
            </div>
            <span class="ms-2">加载知识点中...</span>
        </div>
    `;
    
    try {
        // 从API获取指定章节的知识点
        console.log(`发送API请求: /api/outline/knowledge-points?section=${sectionId}`);
        const response = await fetch(`/api/outline/knowledge-points?section=${sectionId}`);
        if (!response.ok) {
            throw new Error(`获取知识点失败: ${response.statusText}`);
        }
        
        const data = await response.json();
        
        // 添加更详细的调试输出
        console.log(`API返回数据状态: ${response.status}, 完整响应:`, data);
        if (data.section_stats) {
            console.log("章节统计信息:", data.section_stats);
        }
        
        // 确保知识点数据格式正确
        const knowledgePoints = data.knowledge_points || [];
        console.log(`章节 [${sectionId}] 返回的知识点数量: ${knowledgePoints.length}`);
        
        if (knowledgePoints.length > 0) {
            // 输出前5个知识点的详细信息
            console.log("前5个知识点详情:");
            for (let i = 0; i < Math.min(5, knowledgePoints.length); i++) {
                const item = knowledgePoints[i];
                console.log(`知识点 ${i+1}:`, {
                    section: item.section,
                    title: item.title,
                    knowledge: item.knowledge,
                    Knowledge: item.Knowledge, // 添加首字母大写版本的检查
                    difficulty: item.difficulty,
                    tags: item.tags
                });
            }
            
            // 知识点按照章节和子章节分组
            const knowledgePointsBySection = {};
            
            // 处理所有知识点，根据章节和标题分组
            knowledgePoints.forEach(item => {
                // 使用章节编号作为主键 (如 "2.1.2")
                if (!knowledgePointsBySection[item.section]) {
                    knowledgePointsBySection[item.section] = {
                        title: item.title,
                        groups: {}
                    };
                }
                
                // 根据父标签(如果存在)进行二级分组
                const parentTag = item.tags && item.tags.length > 1 ? item.tags[1] : "其他";
                
                if (!knowledgePointsBySection[item.section].groups[parentTag]) {
                    knowledgePointsBySection[item.section].groups[parentTag] = [];
                }
                
                knowledgePointsBySection[item.section].groups[parentTag].push(item);
            });
            
            console.log(`章节 [${sectionId}] 按章节和组分类后的结构:`, knowledgePointsBySection);
            
            // 生成HTML
            let html = '';
            
            // 标记目标章节索引，用于自动滚动
            let targetSectionIndex = -1;
            
            // 处理章节排序
            const sortedSections = Object.keys(knowledgePointsBySection).sort((a, b) => {
                // 按照章节编号排序
                const partsA = a.split('.');
                const partsB = b.split('.');
                
                // 先比较第一个部分
                if (partsA[0] !== partsB[0]) {
                    return parseInt(partsA[0]) - parseInt(partsB[0]);
                }
                
                // 再比较第二个部分
                if (partsA.length > 1 && partsB.length > 1) {
                    if (partsA[1] !== partsB[1]) {
                        return parseInt(partsA[1]) - parseInt(partsB[1]);
                    }
                }
                
                // 最后比较第三个部分
                if (partsA.length > 2 && partsB.length > 2) {
                    return parseInt(partsA[2]) - parseInt(partsB[2]);
                }
                
                return a.localeCompare(b);
            });
            
            // 查找目标章节在排序后的位置
            targetSectionIndex = sortedSections.findIndex(section => section === sectionId);
            console.log(`目标章节 [${sectionId}] 在排序后的位置: ${targetSectionIndex}`);
            
            // 遍历章节
            for (const section of sortedSections) {
                const sectionData = knowledgePointsBySection[section];
                const sectionElementId = `section-${section.replace(/\./g, '-')}`;
                
                // 添加唯一ID以便滚动定位
                html += `<div id="${sectionElementId}" class="knowledge-section mb-3">
                    <div class="knowledge-section-header fw-bold mb-2">${section} ${sectionData.title}</div>
                    <div class="knowledge-section-items">`;
                
                // 处理分组排序，确保"其他"组始终在最后
                const sortedGroups = Object.keys(sectionData.groups).sort((a, b) => {
                    if (a === "其他") return 1;
                    if (b === "其他") return -1;
                    return a.localeCompare(b);
                });
                
                // 遍历章节下的分组
                for (const groupName of sortedGroups) {
                    const items = sectionData.groups[groupName];
                    
                    if (groupName !== "其他" && groupName !== "") {
                        html += `<div class="knowledge-subsection mb-2">
                            <div class="knowledge-subsection-header fst-italic mb-1">${groupName}</div>
                            <div class="ms-2">`;
                    }
                    
                    // 渲染该分组下的所有知识点
                    items.forEach((item, index) => {
                        const itemId = `kp-${section.replace(/\./g, '-')}-${groupName.replace(/\s+/g, '-')}-${index}`;
                        const difficultyBadge = item.difficulty ? 
                            `<span class="badge bg-info ms-1">【${item.difficulty}】</span>` : '';
                        
                        // 检查知识点属性是否存在，兼容不同的属性命名
                        const knowledgeText = item.knowledge || item.Knowledge || "未知知识点";
                        
                        html += `
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" 
                                    name="knowledgePoints" value="${knowledgeText}" 
                                    id="${itemId}">
                                <label class="form-check-label" for="${itemId}">
                                    ${knowledgeText}${difficultyBadge}
                                </label>
                            </div>
                        `;
                    });
                    
                    if (groupName !== "其他" && groupName !== "") {
                        html += `</div></div>`;
                    }
                }
                
                html += `</div></div>`;
            }
            
            // 更新DOM
            console.log(`更新知识点容器HTML, 长度: ${html.length}`);
            knowledgePointsContainer.innerHTML = html || '<div class="alert alert-info">该章节没有知识点数据</div>';
            
            // 自动滚动到选中的章节
            setTimeout(() => {
                scrollToSelectedSection(sectionId, knowledgePointsContainer);
            }, 100);
        } else {
            console.warn(`章节 [${sectionId}] 没有找到知识点，尝试从大纲读取`);
            await loadKnowledgePointsFromOutline(sectionId);
        }
    } catch (error) {
        console.error(`加载知识点失败:`, error);
        await loadKnowledgePointsFromOutline(sectionId);
    }
}

// 滚动到选中的章节
function scrollToSelectedSection(sectionId, container) {
    // 获取目标元素ID
    const targetElementId = `section-${sectionId.replace(/\./g, '-')}`;
    const targetElement = document.getElementById(targetElementId);
    
    if (targetElement) {
        console.log(`找到目标章节元素: #${targetElementId}`);
        
        // 滚动容器到目标元素位置
        const containerTop = container.getBoundingClientRect().top;
        const targetTop = targetElement.getBoundingClientRect().top;
        const scrollTop = targetTop - containerTop + container.scrollTop;
        
        console.log(`滚动容器到位置: ${scrollTop}px`);
        container.scrollTop = scrollTop;
    } else {
        console.warn(`未找到目标章节元素: #${targetElementId}`);
        
        // 尝试使用更宽松的选择器查找最匹配的章节
        const sections = container.querySelectorAll('.knowledge-section');
        let bestMatch = null;
        let bestMatchLength = 0;
        
        sections.forEach(section => {
            const sectionIdAttr = section.id.replace('section-', '').replace(/-/g, '.');
            // 检查是否是目标章节的父级或子级
            if (sectionId.startsWith(sectionIdAttr) || sectionIdAttr.startsWith(sectionId)) {
                // 选择匹配程度最高的（共同前缀最长的）
                const commonPrefixLength = getCommonPrefixLength(sectionId, sectionIdAttr);
                if (commonPrefixLength > bestMatchLength) {
                    bestMatchLength = commonPrefixLength;
                    bestMatch = section;
                }
            }
        });
        
        if (bestMatch) {
            console.log(`找到最匹配的章节元素: #${bestMatch.id}`);
            const containerTop = container.getBoundingClientRect().top;
            const targetTop = bestMatch.getBoundingClientRect().top;
            const scrollTop = targetTop - containerTop + container.scrollTop;
            
            console.log(`滚动容器到最匹配位置: ${scrollTop}px`);
            container.scrollTop = scrollTop;
        } else {
            console.warn(`没有找到任何匹配的章节元素`);
        }
    }
}

// 获取两个字符串的共同前缀长度
function getCommonPrefixLength(str1, str2) {
    const minLength = Math.min(str1.length, str2.length);
    let commonLength = 0;
    
    for (let i = 0; i < minLength; i++) {
        if (str1[i] === str2[i]) {
            commonLength++;
        } else {
            break;
        }
    }
    
    return commonLength;
}

// 直接从大纲文件加载知识点
async function loadKnowledgePointsFromOutline(sectionId) {
    const knowledgePointsContainer = document.getElementById('knowledgePointsContainer');
    if (!knowledgePointsContainer) {
        console.error('知识点容器元素不存在');
        return;
    }
    
    console.log(`尝试从大纲文件直接加载章节 [${sectionId}] 的知识点`);
    
    try {
        // 尝试从本地获取大纲文件内容
        const response = await fetch('/static/data/2024年信息赛大纲.md');
        if (!response.ok) {
            throw new Error('获取大纲文件失败');
        }
        
        const outlineText = await response.text();
        console.log(`大纲文件加载成功，大小: ${outlineText.length} 字节`);
        
        const outlineData = parseOutlineMarkdown(outlineText, sectionId);
        console.log(`成功解析章节 [${sectionId}] 的大纲数据，找到 ${outlineData.length} 个章节数据`);
        
        let html = '';
        
        // 章节排序
        outlineData.sort((a, b) => {
            // 按照章节编号排序
            const partsA = a.id.split('.');
            const partsB = b.id.split('.');
            
            // 先比较第一个部分
            if (partsA[0] !== partsB[0]) {
                return parseInt(partsA[0]) - parseInt(partsB[0]);
            }
            
            // 再比较第二个部分
            if (partsA.length > 1 && partsB.length > 1) {
                if (partsA[1] !== partsB[1]) {
                    return parseInt(partsA[1]) - parseInt(partsB[1]);
                }
            }
            
            // 最后比较第三个部分
            if (partsA.length > 2 && partsB.length > 2) {
                return parseInt(partsA[2]) - parseInt(partsB[2]);
            }
            
            return a.id.localeCompare(b.id);
        });
        
        // 标记目标章节索引
        let targetSectionIndex = outlineData.findIndex(section => section.id === sectionId);
        console.log(`目标章节 [${sectionId}] 在排序后的位置: ${targetSectionIndex}`);
        
        // 渲染章节
        for (const section of outlineData) {
            // 添加唯一ID以便滚动定位
            const sectionElementId = `section-${section.id.replace(/\./g, '-')}`;
            
            html += `<div id="${sectionElementId}" class="knowledge-section mb-3">
                <div class="knowledge-section-header fw-bold mb-2">${section.id} ${section.title}</div>
                <div class="knowledge-section-items">`;
            
            // 分组排序，确保没有标题的分组放在最后
            section.groups.sort((a, b) => {
                if (!a.title && b.title) return 1;
                if (a.title && !b.title) return -1;
                return (a.title || "").localeCompare(b.title || "");
            });
            
            // 渲染分组
            for (const group of section.groups) {
                if (group.title) {
                    html += `<div class="knowledge-subsection mb-2">
                        <div class="knowledge-subsection-header fst-italic mb-1">${group.title}</div>
                        <div class="ms-2">`;
                }
                
                // 渲染知识点
                group.items.forEach((item, index) => {
                    const itemId = `kp-${section.id.replace(/\./g, '-')}-${index}`;
                    const difficultyBadge = item.difficulty ? 
                        `<span class="badge bg-info ms-1">【${item.difficulty}】</span>` : '';
                    
                    html += `
                        <div class="form-check">
                            <input class="form-check-input" type="checkbox" 
                                name="knowledgePoints" value="${item.text}" 
                                id="${itemId}">
                            <label class="form-check-label" for="${itemId}">
                                ${item.text}${difficultyBadge}
                            </label>
                        </div>
                    `;
                });
                
                if (group.title) {
                    html += `</div></div>`;
                }
            }
            
            html += `</div></div>`;
        }
        
        console.log(`成功生成章节 [${sectionId}] 的知识点HTML，长度: ${html.length}`);
        knowledgePointsContainer.innerHTML = html || '<div class="alert alert-info">该章节没有知识点数据</div>';
        
        // 自动滚动到选中的章节
        setTimeout(() => {
            scrollToSelectedSection(sectionId, knowledgePointsContainer);
        }, 100);
        
    } catch (error) {
        console.error(`从大纲加载章节 [${sectionId}] 知识点失败:`, error);
        renderDefaultKnowledgePointsFromOutline(sectionId);
    }
}

// 解析大纲Markdown文件
function parseOutlineMarkdown(markdownText, targetSection) {
    console.log(`解析大纲文件，目标章节: [${targetSection}]`);
    
    const lines = markdownText.split('\n');
    const result = [];
    
    let currentH2 = null;
    let currentH3 = null;
    let currentGroup = null;
    let inTargetSection = false;
    
    // 二级和三级标题的正则
    const h2Regex = /^## ([\d\.]+) (.+)$/;
    const h3Regex = /^### ([\d\.]+) (.+)$/;
    // 列表项的正则
    const numberedItemRegex = /^(\d+)\. \【(\d+)\】(.+)$/;
    const bulletItemRegex = /^   - \【(\d+)\】(.+)$/;
    
    // 判断章节是否是目标章节的同级或下级
    function isSameOrChildSection(sectionId, targetId) {
        // 如果是完全匹配，直接返回true
        if (sectionId === targetId) {
            console.log(`章节 [${sectionId}] 与目标章节 [${targetId}] 完全匹配`);
            return true;
        }
        
        // 判断是否是上级章节（例如 targetId="2.1", sectionId="2"）
        if (targetId.startsWith(sectionId + ".")) {
            console.log(`章节 [${sectionId}] 是目标章节 [${targetId}] 的上级章节`);
            return true;
        }
        
        // 判断是否是下级章节（例如 targetId="2", sectionId="2.1"）
        if (sectionId.startsWith(targetId + ".")) {
            console.log(`章节 [${sectionId}] 是目标章节 [${targetId}] 的下级章节`);
            return true;
        }
        
        // 判断是否是同级章节的子章节
        // 例如，如果 targetId="2.1", 那么 "2.2.1" 等同级章节的子章节也应该显示
        const targetParts = targetId.split('.');
        const sectionParts = sectionId.split('.');
        
        // 如果是根章节，保留所有内容
        if (targetId === 'all') {
            console.log(`目标章节是"all"，章节 [${sectionId}] 应显示`);
            return true;
        }
        
        // 对于"2.1"这样的格式，我们需要检查是否同属于"2"这个大章节
        if (targetParts.length >= 2 && sectionParts.length >= 2) {
            // 检查前面部分是否相同（例如 "2.1" 和 "2.2" 同属于 "2"）
            if (targetParts[0] === sectionParts[0]) {
                console.log(`章节 [${sectionId}] 与目标章节 [${targetId}] 同属于第 ${targetParts[0]} 大章节`);
                
                // 特殊处理2.1.2章节匹配
                if (targetId === "2.1.2" && sectionId === "2.1.2") {
                    console.log(`特殊匹配: 目标章节和当前章节都是2.1.2`);
                    return true;
                }
                
                // 对于同级章节，只要同属一个父章节，就应该显示
                return true;
            }
        }
        
        return false;
    }
    
    let matchedSectionCount = 0;
    
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();
        if (!line) continue;
        
        // 匹配二级标题 (## 2.1 入门级)
        const h2Match = line.match(h2Regex);
        if (h2Match) {
            const id = h2Match[1];
            const title = h2Match[2];
            
            // 检查是否为目标章节或相关章节
            if (isSameOrChildSection(id, targetSection)) {
                inTargetSection = true;
                currentH2 = { id, title, groups: [] };
                result.push(currentH2);
                console.log(`匹配到相关的二级标题: ${id} ${title}`);
                matchedSectionCount++;
            } else {
                inTargetSection = false;
            }
            continue;
        }
        
        if (!inTargetSection) continue;
        
        // 匹配三级标题 (### 2.1.1 基础知识与编程环境)
        const h3Match = line.match(h3Regex);
        if (h3Match) {
            const id = h3Match[1];
            const title = h3Match[2];
            
            // 检查是否为目标章节或相关章节
            if (isSameOrChildSection(id, targetSection)) {
                currentH3 = { id, title, groups: [] };
                result.push(currentH3);
                currentGroup = null;
                console.log(`匹配到相关的三级标题: ${id} ${title}`);
                matchedSectionCount++;
            } else {
                currentH3 = null; // 不相关的章节，不解析其内容
            }
            continue;
        }
        
        // 不是我们要找的章节就跳过
        if (!currentH3) continue;
        
        // 数字列表项，可能是一个新组 (1. 【1】计算机的基本构成)
        const numberedMatch = line.match(numberedItemRegex);
        if (numberedMatch) {
            const number = numberedMatch[1];
            const difficulty = numberedMatch[2];
            const text = numberedMatch[3].trim();
            
            // 检查下一行是否有缩进的列表项，如果有则这是一个组标题
            if (i + 1 < lines.length && lines[i + 1].trim().startsWith('   -')) {
                currentGroup = { title: text, items: [] };
                currentH3.groups.push(currentGroup);
                console.log(`检测到组标题: ${text} (在 ${currentH3.id} 下)`);
            } else {
                // 这是一个直接的知识点
                if (!currentGroup) {
                    currentGroup = { title: '', items: [] };
                    currentH3.groups.push(currentGroup);
                }
                currentGroup.items.push({ text, difficulty });
                
                // 如果是特殊章节2.1.2，添加日志
                if (currentH3.id === "2.1.2") {
                    console.log(`在2.1.2章节下找到独立知识点: ${text}, 难度: ${difficulty}`);
                }
            }
            continue;
        }
        
        // 缩进的列表项 (   - 【1】标识符、关键字、常量、变量、字符串、表达式的概念)
        const bulletMatch = line.match(bulletItemRegex);
        if (bulletMatch && currentGroup) {
            const difficulty = bulletMatch[1];
            const text = bulletMatch[2].trim();
            currentGroup.items.push({ text, difficulty });
            
            // 如果是特殊章节2.1.2，添加日志
            if (currentH3 && currentH3.id === "2.1.2") {
                console.log(`在2.1.2章节中找到子条目: ${text}, 难度: ${difficulty}`);
            }
        }
    }
    
    console.log(`解析大纲文件完成，匹配到 ${matchedSectionCount} 个相关章节，共 ${result.length} 个章节数据`);
    
    return result;
}

// 为特定章节渲染默认知识点（基于大纲文件格式）
function renderDefaultKnowledgePointsFromOutline(sectionId) {
    const knowledgePointsContainer = document.getElementById('knowledgePointsContainer');
    if (!knowledgePointsContainer) {
        console.error('知识点容器元素不存在');
        return;
    }
    
    console.log(`渲染章节 [${sectionId}] 的默认知识点`);

    // 创建一个简单的模拟数据结构
    const mockData = [];
    
    // 根据章节ID创建不同的模拟数据
    if (sectionId.startsWith("2.1")) {
        // 入门级
        mockData.push({
            id: sectionId,
            title: "入门级章节内容",
            groups: [
                {
                    title: "基础知识点",
                    items: [
                        { text: "这是章节 " + sectionId + " 的示例知识点1", difficulty: "1" },
                        { text: "这是章节 " + sectionId + " 的示例知识点2", difficulty: "2" }
                    ]
                }
            ]
        });
    } else if (sectionId.startsWith("2.2")) {
        // 提高级
        mockData.push({
            id: sectionId,
            title: "提高级章节内容",
            groups: [
                {
                    title: "进阶知识点",
                    items: [
                        { text: "这是章节 " + sectionId + " 的示例知识点1", difficulty: "3" },
                        { text: "这是章节 " + sectionId + " 的示例知识点2", difficulty: "4" }
                    ]
                }
            ]
        });
    } else if (sectionId.startsWith("2.3")) {
        // NOI级
        mockData.push({
            id: sectionId,
            title: "NOI级章节内容",
            groups: [
                {
                    title: "高级知识点",
                    items: [
                        { text: "这是章节 " + sectionId + " 的示例知识点1", difficulty: "7" },
                        { text: "这是章节 " + sectionId + " 的示例知识点2", difficulty: "8" }
                    ]
                }
            ]
        });
    } else {
        // 通用
        mockData.push({
            id: sectionId,
            title: "章节内容",
            groups: [
                {
                    title: "一般知识点",
                    items: [
                        { text: "无法获取章节 " + sectionId + " 的实际知识点", difficulty: "1" },
                        { text: "请尝试刷新或选择其他章节", difficulty: "1" }
                    ]
                }
            ]
        });
    }
    
    let html = '';
    
    // 渲染模拟数据
    for (const section of mockData) {
        // 添加唯一ID以便滚动定位
        const sectionElementId = `section-${section.id.replace(/\./g, '-')}`;
        
        html += `<div id="${sectionElementId}" class="knowledge-section mb-3">
            <div class="knowledge-section-header fw-bold mb-2">${section.id} ${section.title}</div>
            <div class="knowledge-section-items alert alert-warning">
                <p>⚠️ 无法加载实际知识点，显示默认内容</p>
                <div class="knowledge-section-items">`;
        
        for (const group of section.groups) {
            if (group.title) {
                html += `<div class="knowledge-subsection mb-2">
                    <div class="knowledge-subsection-header fst-italic mb-1">${group.title}</div>
                    <div class="ms-2">`;
            }
            
            group.items.forEach((item, index) => {
                const itemId = `kp-default-${section.id.replace(/\./g, '-')}-${index}`;
                const difficultyBadge = item.difficulty ? 
                    `<span class="badge bg-info ms-1">【${item.difficulty}】</span>` : '';
                
                html += `
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" 
                            name="knowledgePoints" value="${item.text}" 
                            id="${itemId}">
                        <label class="form-check-label" for="${itemId}">
                            ${item.text}${difficultyBadge}
                        </label>
                    </div>
                `;
            });
            
            if (group.title) {
                html += `</div></div>`;
            }
        }
        
        html += `</div></div>`;
    }
    
    knowledgePointsContainer.innerHTML = html;
    
    // 自动滚动到选中的章节
    setTimeout(() => {
        scrollToSelectedSection(sectionId, knowledgePointsContainer);
    }, 100);
}

// 生成AI题目
async function generateAIProblem(retryCount = 0) {
    const MAX_RETRIES = 2; // 最多重试2次（总共最多尝试3次）
    
    // 获取表单数据
    const aiModel = document.getElementById('aiModelSelect').value;
    const title = document.getElementById('generateTitle').value;
    const outlineSection = document.getElementById('outlineSection').value;
    const difficultySelect = document.getElementById('generateDifficulty');
    const difficulty = difficultySelect.value;
    
    // 获取选中的知识点
    const knowledgePoints = [];
    document.querySelectorAll('input[name="knowledgePoints"]:checked').forEach(item => {
        knowledgePoints.push(item.value);
    });
    
    // 获取题目类型
    const problemType = document.querySelector('input[name="problemType"]:checked').value;
    
    // 其他设置
    const timeComplexity = document.getElementById('timeComplexity').value;
    const spaceComplexity = document.getElementById('spaceComplexity').value;
    const additionalReqs = document.getElementById('additionalReqs').value;
    const testCaseCount = parseInt(document.getElementById('testCaseCount').value) || 5;
    
    // 强制设置为true，确保总是生成参考解答和思维分析
    const includeReferenceSolution = true;
    const includeAnalysis = true;
    
    // 验证
    if (knowledgePoints.length === 0) {
        alert('请至少选择一个知识点');
        return;
    }
    
    // 准备请求数据
    const requestData = {
        title: title,
        outline_section: outlineSection,
        knowledge_points: knowledgePoints,
        difficulty: difficulty,
        problem_type: problemType,
        time_complexity: timeComplexity,
        space_complexity: spaceComplexity,
        additional_reqs: additionalReqs,
        test_case_count: testCaseCount,
        include_reference_solution: includeReferenceSolution,
        include_analysis: includeAnalysis,
        model_type: aiModel
    };
    
    console.log('生成题目请求数据:', requestData);
    
    // 显示进度，隐藏表单
    document.getElementById('generateAIProblemForm').style.display = 'none';
    document.getElementById('generateProgress').style.display = 'block';
    document.getElementById('generateModalFooter').style.display = 'none';

    // 更新进度信息
    const progressText = document.querySelector('#generateProgress p');
    if (retryCount > 0) {
        progressText.textContent = `AI正在生成题目，请稍候...（第${retryCount + 1}次尝试）`;
    } else {
        progressText.textContent = `AI正在生成题目，请稍候...`;
    }
    
    try {
        // 设置超时时间为2分钟
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 120000);
        
        const response = await fetch('/api/problems/generate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestData),
            signal: controller.signal
        });
        
        // 清除超时
        clearTimeout(timeoutId);
        
        if (!response.ok) {
            throw new Error(`服务器响应错误: ${response.status} ${response.statusText}`);
        }
        
        const generatedProblem = await response.json();
        console.log('生成的题目:', generatedProblem);
        
        // 存储生成的题目到全局变量，以便后续保存使用
        window.generatedProblem = generatedProblem;
        
        // 显示生成结果
        displayGeneratedProblem(generatedProblem);
    } catch (error) {
        console.error('生成题目失败:', error);
        
        // 处理不同类型的错误
        let errorMessage = '';
        let canRetry = false;
        
        if (error.name === 'AbortError') {
            errorMessage = '生成题目请求超时，可能是服务器处理时间过长或网络问题';
            canRetry = true;
        } else if (error.message.includes('reset by peer') || 
                   error.message.includes('network') || 
                   error.message.includes('connection') ||
                   error.message.includes('timeout')) {
            errorMessage = `网络连接错误: ${error.message}`;
            canRetry = true;
        } else {
            errorMessage = `生成题目失败: ${error.message}`;
        }
        
        // 如果可以重试且未超过最大重试次数
        if (canRetry && retryCount < MAX_RETRIES) {
            console.log(`正在进行第${retryCount + 1}次重试...`);
            // 显示正在重试的消息
            document.querySelector('#generateProgress p').textContent = 
                `检测到网络问题，正在重试 (${retryCount + 1}/${MAX_RETRIES})...`;
            
            // 延迟2秒后重试
            setTimeout(() => generateAIProblem(retryCount + 1), 2000);
            return;
        }
        
        // 超过重试次数或不能重试的错误
        alert(errorMessage);
        
        // 恢复表单显示
        document.getElementById('generateProgress').style.display = 'none';
        document.getElementById('generateAIProblemForm').style.display = 'block';
        document.getElementById('generateModalFooter').style.display = 'flex';
    }
}

// 显示生成的题目
function displayGeneratedProblem(problem) {
    // 隐藏加载中显示，显示结果
    document.getElementById('generateProgress').style.display = 'none';
    document.getElementById('generatedProblemResult').style.display = 'block';
    
    console.log('显示生成的题目:', problem);
    
    // 检查并设置题目信息
    document.getElementById('generatedProblemTitle').textContent = problem.title || '未命名题目';
    document.getElementById('previewTitle').textContent = problem.title || '未命名题目';
    document.getElementById('previewDifficulty').textContent = problem.difficulty || 'Medium';
    document.getElementById('previewLimits').textContent = `时间限制: ${problem.time_limit || 1000}ms, 内存限制: ${problem.memory_limit || (256 * 1024)}KB`;
    
    // 确保描述存在并正确显示
    if (problem.description && problem.description.trim() !== '') {
        document.getElementById('previewDescription').innerHTML = problem.description.replace(/\n/g, '<br>');
    } else {
        document.getElementById('previewDescription').innerHTML = '<span class="text-warning">注意: 题目描述为空，请手动添加描述。</span>';
    }
    
    // 示例
    let examplesHtml = '';
    if (problem.test_cases && problem.test_cases.length > 0) {
        problem.test_cases.forEach((tc, index) => {
            if (tc.is_example) {
                examplesHtml += `
                    <div class="card mb-2">
                        <div class="card-header">样例 ${index + 1}</div>
                        <div class="card-body">
                            <div class="row">
                                <div class="col-md-6">
                                    <strong>输入:</strong>
                                    <pre>${tc.input}</pre>
                                </div>
                                <div class="col-md-6">
                                    <strong>输出:</strong>
                                    <pre>${tc.output}</pre>
                                </div>
                            </div>
                        </div>
                    </div>
                `;
            }
        });
    }
    
    if (examplesHtml === '') {
        examplesHtml = '<div class="alert alert-warning">无示例测试用例</div>';
    }
    
    document.getElementById('previewExamples').innerHTML = examplesHtml;
    
    // 知识点标签
    let tagsHtml = '';
    if (problem.knowledge_tag && problem.knowledge_tag.length > 0) {
        tagsHtml = problem.knowledge_tag.map(tag => `<span class="badge bg-secondary me-1">${tag}</span>`).join('');
    } else {
        tagsHtml = '<span class="text-muted">无知识点标签</span>';
    }
    document.getElementById('previewTags').innerHTML = tagsHtml;
    
    // 参考解答和思维分析 - 添加可折叠的部分
    const previewContainer = document.querySelector('#generatedProblemResult .card-body');
    
    // 移除旧的参考解答和思维分析块（如果有）
    const oldSolution = document.getElementById('previewReferenceSolution');
    const oldAnalysis = document.getElementById('previewThinkingAnalysis');
    if (oldSolution) oldSolution.remove();
    if (oldAnalysis) oldAnalysis.remove();
    
    // 添加参考解答
    if (problem.reference_solution && problem.reference_solution.trim() !== '') {
        const solutionDiv = document.createElement('div');
        solutionDiv.id = 'previewReferenceSolution';
        solutionDiv.className = 'mt-4';
        solutionDiv.innerHTML = `
            <h6>参考解答:</h6>
            <div class="mb-3">
                <button class="btn btn-outline-primary w-100 text-start" type="button" 
                        data-bs-toggle="collapse" data-bs-target="#previewCollapseSolution" 
                        aria-expanded="false" aria-controls="previewCollapseSolution">
                    <i class="bi bi-code-square"></i> 参考解答 <span class="small text-muted">(点击展开)</span>
                </button>
                <div class="collapse mt-2" id="previewCollapseSolution">
                    <div class="card card-body">
                        <pre>${problem.reference_solution.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</pre>
                    </div>
                </div>
            </div>
        `;
        previewContainer.appendChild(solutionDiv);
    }
    
    // 添加思维分析
    if (problem.thinking_analysis && problem.thinking_analysis.trim() !== '') {
        const analysisDiv = document.createElement('div');
        analysisDiv.id = 'previewThinkingAnalysis';
        analysisDiv.className = 'mt-3';
        analysisDiv.innerHTML = `
            <h6>思维分析:</h6>
            <div class="mb-3">
                <button class="btn btn-outline-info w-100 text-start" type="button" 
                        data-bs-toggle="collapse" data-bs-target="#previewCollapseAnalysis" 
                        aria-expanded="false" aria-controls="previewCollapseAnalysis">
                    <i class="bi bi-lightbulb"></i> 思维分析 <span class="small text-muted">(点击展开)</span>
                </button>
                <div class="collapse mt-2" id="previewCollapseAnalysis">
                    <div class="card card-body">
                        ${problem.thinking_analysis.replace(/\n/g, '<br>')}
                    </div>
                </div>
            </div>
        `;
        previewContainer.appendChild(analysisDiv);
    }
    
    // 启用保存按钮
    document.getElementById('saveGeneratedProblemBtn').disabled = false;
    document.getElementById('generateModalFooter').style.display = 'flex';
}

// 保存生成的题目
async function saveGeneratedProblem() {
    if (!window.generatedProblem) {
        alert('没有可保存的题目');
        return;
    }
    
    try {
        const response = await fetch('/api/problems/save-generated', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(window.generatedProblem)
        });
        
        if (!response.ok) {
            throw new Error(`保存题目失败: ${response.statusText}`);
        }
        
        const result = await response.json();
        console.log('保存题目结果:', result);
        
        // 关闭模态框
        const generateModal = bootstrap.Modal.getInstance(document.getElementById('generateAIProblemModal'));
        if (generateModal) {
            generateModal.hide();
        }
        
        // 刷新题目列表
        loadProblems();
        
        // 显示成功消息
        alert(`题目"${window.generatedProblem.title}"已成功保存到题库！`);
    } catch (error) {
        console.error('保存题目失败:', error);
        alert(`保存题目失败: ${error.message}`);
    }
}

// 调试辅助函数：记录问题详情的关键属性
function logProblemDetails(problem, examples) {
    console.log('问题ID:', problem.id);
    console.log('问题标题:', problem.title);
    console.log('问题难度:', problem.difficulty);
    console.log('问题描述长度:', problem.description ? problem.description.length : 0);
    console.log('样例数量:', examples ? examples.length : 0);
    
    if (examples && examples.length > 0) {
        console.log('第一个样例输入:', examples[0].input);
        console.log('第一个样例输出:', examples[0].output);
    }
    
    console.log('知识点标签:', problem.knowledge_tag);
}