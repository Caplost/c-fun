// DOM Elements
const problemsList = document.getElementById('problems-list');
const problemDetail = document.getElementById('problem-detail');
const submissionResult = document.getElementById('submission-result');
const problemsContainer = document.querySelector('.problems-container');
const tagFilterContainer = document.getElementById('tag-filter-container');
const filterTabs = document.querySelectorAll('.filter-tab');
const problemTitle = document.getElementById('problem-title');
const problemDifficulty = document.getElementById('problem-difficulty');
const problemLimits = document.getElementById('problem-limits');
const problemDescriptionText = document.getElementById('problem-description-text');
const problemInput = document.getElementById('problem-input');
const problemOutput = document.getElementById('problem-output');
const examplesContainer = document.getElementById('examples-container');
const codeEditor = document.getElementById('code-editor');
const submitBtn = document.getElementById('submit-btn');
const resultStatus = document.getElementById('result-status');
const resultDetails = document.getElementById('result-details');
const backToProblemBtn = document.getElementById('back-to-problem');
const importProblemsBtn = document.getElementById('import-problems-btn');
const importDialog = document.getElementById('import-dialog');
const closeDialogBtn = document.querySelector('.close-btn');
const cancelImportBtn = document.getElementById('cancel-import');
const submitImportBtn = document.getElementById('submit-import');
const importJsonTextarea = document.getElementById('import-json');

// 难度翻译
const difficultyTranslation = {
    'Easy': '简单',
    'Medium': '中等',
    'Hard': '困难'
};

// 状态翻译
const statusTranslation = {
    'Accepted': '通过',
    'Wrong Answer': '答案错误',
    'Compilation Error': '编译错误',
    'Runtime Error': '运行时错误',
    'Time Limit Exceeded': '超时',
    'Internal Error': '内部错误',
    'Pending': '评测中'
};

// Current state
let currentProblemId = null;
let currentUser = { id: 1 }; // Mock user for demo purposes
let allProblems = []; // 存储所有题目数据
let activeFilter = 'all'; // 当前激活的过滤选项：all, solved, unsolved
let activeTags = []; // 当前激活的标签过滤

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    loadProblems();
    setupEventListeners();

    // Add a default template code to the editor
    codeEditor.value = `#include <iostream>
using namespace std;

int main() {
    // 在这里编写你的解答
    
    return 0;
}`;
});

// Setup event listeners
function setupEventListeners() {
    // Submit Button
    if (submitBtn) {
        submitBtn.addEventListener('click', submitSolution);
    }

    // Back to problem button
    if (backToProblemBtn) {
        backToProblemBtn.addEventListener('click', () => {
            submissionResult.style.display = 'none';
            problemDetail.style.display = 'block';
        });
    }

    // Back to problems list button
    const backToProblemsBtn = document.getElementById('back-to-problems-btn');
    if (backToProblemsBtn) {
        backToProblemsBtn.addEventListener('click', goBackToProblems);
    }

    // Import Problems Button
    if (importProblemsBtn) {
        importProblemsBtn.addEventListener('click', openImportDialog);
    }

    // Close import dialog
    if (closeDialogBtn) {
        closeDialogBtn.addEventListener('click', closeImportDialog);
    }

    // Cancel import
    if (cancelImportBtn) {
        cancelImportBtn.addEventListener('click', closeImportDialog);
    }

    // Submit import
    if (submitImportBtn) {
        submitImportBtn.addEventListener('click', submitImportProblems);
    }

    // Navigation
    document.getElementById('nav-problems').addEventListener('click', (e) => {
        e.preventDefault();
        goBackToProblems();
    });
    
    document.getElementById('nav-submissions').addEventListener('click', (e) => {
        e.preventDefault();
        alert('功能开发中，敬请期待！');
    });
    
    // 点击对话框外部时关闭
    window.addEventListener('click', (e) => {
        if (e.target === importDialog) {
            closeImportDialog();
        }
    });
    
    // 设置过滤Tab点击事件
    filterTabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const filter = tab.getAttribute('data-filter');
            activeFilter = filter;
            
            // 更新激活状态
            filterTabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            
            // 应用过滤
            applyFilters();
        });
    });
}

// Load problems
async function loadProblems() {
    try {
        // Check if main.js has already loaded problems
        const problemsList = document.getElementById('problemsList');
        if (problemsList && problemsList.children.length > 0) {
            // Problems already loaded by main.js, no need to load again
            return;
        }

        const userID = currentUser.id;
        const response = await fetch(`/api/problems/status?user_id=${userID}`);
        const data = await response.json();
        
        allProblems = data;
        
        // Extract all tags
        const allTags = extractAllTags(data);
        renderTagFilters(allTags);
        
        // Apply filters
        applyFilters();
    } catch (error) {
        console.error('Error loading problems:', error);
        alert('加载题目列表时出错，请刷新页面重试。');
    }
}

// 提取所有唯一的知识点标签
function extractAllTags(problems) {
    const tagsSet = new Set();
    problems.forEach(problem => {
        if (problem.knowledge_tag && Array.isArray(problem.knowledge_tag)) {
            problem.knowledge_tag.forEach(tag => tagsSet.add(tag));
        }
    });
    return Array.from(tagsSet);
}

// 渲染标签过滤器
function renderTagFilters(tags) {
    if (!tags || tags.length === 0) {
        tagFilterContainer.innerHTML = '<span class="empty-tags">暂无标签</span>';
        return;
    }
    
    // 清除现有标签
    tagFilterContainer.innerHTML = '';
    
    // 添加"清除筛选"按钮 - 仅当有激活的标签时显示
    const clearBtn = document.createElement('span');
    clearBtn.className = 'tag-badge clear-filter';
    clearBtn.textContent = '清除筛选';
    clearBtn.style.display = activeTags.length > 0 ? 'inline-block' : 'none';
    clearBtn.addEventListener('click', () => {
        // 清除所有选中的标签
        activeTags = [];
        document.querySelectorAll('.tag-badge').forEach(b => b.classList.remove('active'));
        clearBtn.style.display = 'none';
        applyFilters();
    });
    tagFilterContainer.appendChild(clearBtn);
    
    // 添加标签
    tags.forEach(tag => {
        const badge = document.createElement('span');
        badge.className = 'tag-badge';
        badge.textContent = tag;
        badge.dataset.tag = tag;
        
        // 如果该标签已激活，添加active类
        if (activeTags.includes(tag)) {
            badge.classList.add('active');
        }
        
        badge.addEventListener('click', () => {
            const tagText = badge.dataset.tag;
            
            if (badge.classList.contains('active')) {
                // 取消选中
                badge.classList.remove('active');
                activeTags = activeTags.filter(t => t !== tagText);
            } else {
                // 选中
                badge.classList.add('active');
                activeTags.push(tagText);
            }
            
            // 更新清除按钮显示状态
            clearBtn.style.display = activeTags.length > 0 ? 'inline-block' : 'none';
            
            // 应用过滤
            applyFilters();
        });
        
        tagFilterContainer.appendChild(badge);
    });
}

// 应用过滤器
function applyFilters() {
    if (!allProblems || allProblems.length === 0) return;
    
    // 根据解题状态和标签过滤
    const filteredProblems = allProblems.filter(problem => {
        // 解题状态过滤
        if (activeFilter === 'solved' && !problem.solved) return false;
        if (activeFilter === 'unsolved' && problem.solved) return false;
        
        // 标签过滤
        if (activeTags.length > 0) {
            // 如果没有知识点标签，则跳过
            if (!problem.knowledge_tag || !Array.isArray(problem.knowledge_tag) || problem.knowledge_tag.length === 0) {
                return false;
            }
            
            // 检查是否包含任何一个选中的标签
            const hasTag = activeTags.some(tag => problem.knowledge_tag.includes(tag));
            if (!hasTag) return false;
        }
        
        return true;
    });
    
    // 按照ID从小到大排序
    filteredProblems.sort((a, b) => a.id - b.id);
    
    // 渲染过滤后的题目列表
    renderProblemsList(filteredProblems);
}

// Render problems list
function renderProblemsList(problems) {
    const problemsList = document.getElementById('problemsList');
    if (!problemsList) return;
    
    problemsList.innerHTML = '';

    problems.forEach(problem => {
        const statusClass = getStatusClass(problem);
        const statusText = getStatusText(problem);
        
        const row = document.createElement('tr');
        row.className = 'problem-card';
        row.setAttribute('data-id', problem.id);
        
        row.innerHTML = `
            <td>${problem.id}</td>
            <td>
                <a href="#problem-${problem.id}" class="problem-link">
                    ${problem.title}
                </a>
            </td>
            <td>${difficultyTranslation[problem.difficulty] || problem.difficulty}</td>
            <td><span class="status-indicator ${statusClass}"></span> ${statusText}</td>
        `;
        
        // Add click event
        row.addEventListener('click', () => {
            loadProblemDetail(problem.id);
        });
        
        problemsList.appendChild(row);
    });
}

// Load problem detail
async function loadProblemDetail(problemId) {
    try {
        currentProblemId = problemId;
        
        // 使用带状态的API获取问题详情
        const response = await fetch(`/api/problems/${problemId}/status?user_id=${currentUser.id}`);
        const data = await response.json();
        
        // 调试输出问题详情，特别是描述字段
        console.log('Problem detail loaded:', data);
        console.log('Problem description:', data.problem.description);
        
        renderProblemDetail(data.problem, data.examples, data);
        
        // Show problem detail, hide problems list
        problemsList.style.display = 'none';
        problemDetail.style.display = 'block';
    } catch (error) {
        console.error('Error loading problem detail:', error);
        alert('加载题目详情时出错，请重试。');
    }
}

// Render problem detail
function renderProblemDetail(problem, examples, statusData) {
    // Set problem details
    problemTitle.textContent = problem.title;
    const translatedDifficulty = difficultyTranslation[problem.difficulty] || problem.difficulty;
    problemDifficulty.textContent = translatedDifficulty;
    
    // 添加知识点标签和状态信息
    let metaInfo = `时间限制: ${problem.time_limit}ms | 内存限制: ${problem.memory_limit / 1000}MB`;
    
    // 添加状态信息
    if (statusData.attempted) {
        metaInfo += ` | ${statusData.solved ? '已解决' : '尝试中'}`;
        if (statusData.failed_attempts > 0) {
            metaInfo += ` (失败: ${statusData.failed_attempts}次)`;
        }
    }
    
    problemLimits.textContent = metaInfo;
    
    // 添加知识点标签显示
    if (problem.knowledge_tag && problem.knowledge_tag.length > 0) {
        const tagsHtml = `
            <div class="problem-knowledge-tags">
                <h3>知识点标签:</h3>
                <div class="tags-container">
                    ${problem.knowledge_tag.map(tag => `<span class="problem-tag">${tag}</span>`).join('')}
                </div>
            </div>
        `;
        // 在题目描述前插入标签区域
        const tagsElement = document.createElement('div');
        tagsElement.innerHTML = tagsHtml;
        problemDescriptionText.parentNode.insertBefore(tagsElement, problemDescriptionText.parentNode.firstChild);
    }
    
    // 解析描述中的各个部分
    const descriptionParts = parseProblemDescription(problem.description);
    
    // 设置题目描述、输入、输出
    try {
        problemDescriptionText.innerHTML = descriptionParts.description || '无题目描述';
        problemInput.innerHTML = descriptionParts.input || '';
        problemOutput.innerHTML = descriptionParts.output || '';
        
        // 如果描述为空，显示提示信息
        if (!descriptionParts.description || descriptionParts.description.trim() === '') {
            problemDescriptionText.innerHTML = '<p class="text-warning">题目描述不可用或为空</p>';
        }
    } catch (error) {
        console.error('Error setting problem description:', error);
        problemDescriptionText.innerHTML = '<p class="text-danger">显示题目描述时出错</p>';
    }
    
    // Render examples
    let examplesHtml = '';
    examples.forEach((example, index) => {
        examplesHtml += `
            <div class="example">
                <div class="example-header">样例 ${index + 1}</div>
                <div class="example-content">
                    <strong>输入:</strong>
                    <div class="example-input">${example.input}</div>
                    <strong>输出:</strong>
                    <div class="example-output">${example.output}</div>
                </div>
            </div>
        `;
    });
    
    examplesContainer.innerHTML = examplesHtml || '<div class="empty">此题目没有提供样例。</div>';
}

// 解析题目描述，将其分为题目描述、输入和输出部分
function parseProblemDescription(htmlDescription) {
    if (!htmlDescription) {
        return {
            description: '题目描述不可用',
            input: '',
            output: ''
        };
    }
    
    const result = {
        description: '',
        input: '',
        output: ''
    };
    
    // 查找【题目描述】部分
    const descriptionMatch = htmlDescription.match(/<h3>【题目描述】<\/h3>(.*?)(?=<h3>|$)/s);
    if (descriptionMatch && descriptionMatch[1]) {
        result.description = descriptionMatch[1].trim();
    }
    
    // 查找输入部分
    const inputMatch = htmlDescription.match(/<h3>输入<\/h3>(.*?)(?=<h3>|$)/s);
    if (inputMatch && inputMatch[1]) {
        result.input = inputMatch[1].trim();
    }
    
    // 查找输出部分
    const outputMatch = htmlDescription.match(/<h3>输出<\/h3>(.*?)(?=<h3>|$)/s);
    if (outputMatch && outputMatch[1]) {
        result.output = outputMatch[1].trim();
    }
    
    // 如果没有找到匹配的结构，使用整个描述
    if (!result.description && !result.input && !result.output) {
        result.description = htmlDescription;
    }
    
    return result;
}

// Submit solution
async function submitSolution() {
    if (!currentProblemId) return;
    
    const code = codeEditor.value.trim();
    if (!code) {
        alert('请先编写您的代码再提交。');
        return;
    }
    
    // Disable submit button
    submitBtn.disabled = true;
    submitBtn.textContent = '提交中...';
    
    try {
        const response = await fetch(`/api/problems/${currentProblemId}/submissions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                user_id: currentUser.id,
                code: code,
                language: 'cpp'
            })
        });
        
        const data = await response.json();
        
        // Poll for submission result
        pollSubmissionResult(data.submission_id);
    } catch (error) {
        console.error('Error submitting solution:', error);
        alert('提交解答时出错，请重试。');
        
        // Re-enable submit button
        submitBtn.disabled = false;
        submitBtn.textContent = '提交';
    }
}

// Poll for submission result
async function pollSubmissionResult(submissionId) {
    try {
        const response = await fetch(`/api/submissions/${submissionId}`);
        const data = await response.json();
        
        if (data.submission.status === 'Pending') {
            // If still pending, poll again after a delay
            setTimeout(() => pollSubmissionResult(submissionId), 1000);
            return;
        }
        
        // Show result
        renderSubmissionResult(data.submission, data.test_results);
        
        // Re-enable submit button
        submitBtn.disabled = false;
        submitBtn.textContent = '提交';
    } catch (error) {
        console.error('Error polling submission result:', error);
        alert('获取提交结果时出错，请重试。');
        
        // Re-enable submit button
        submitBtn.disabled = false;
        submitBtn.textContent = '提交';
    }
}

// Render submission result
function renderSubmissionResult(submission, testResults) {
    // Set status
    const translatedStatus = statusTranslation[submission.status] || submission.status;
    resultStatus.textContent = translatedStatus;
    resultStatus.className = 'status status-' + submission.status.toLowerCase().replace(' ', '-');
    
    // Set details
    let detailsHtml = `
        <div>
            <strong>运行时间:</strong> ${submission.run_time}ms
            <strong>内存:</strong> ${submission.memory}KB
        </div>
    `;
    
    // Show test results
    if (testResults && testResults.length > 0) {
        detailsHtml += '<h3>测试结果</h3>';
        
        testResults.forEach((result, index) => {
            const translatedResultStatus = statusTranslation[result.status] || result.status;
            detailsHtml += `
                <div class="test-case-result">
                    <h4>测试用例 ${index + 1}: ${translatedResultStatus}</h4>
                    ${result.status !== 'Accepted' ? `
                        <div>
                            <strong>你的输出:</strong>
                            <div class="test-output">${result.output}</div>
                        </div>
                    ` : ''}
                </div>
            `;
        });
    }
    
    resultDetails.innerHTML = detailsHtml;
    
    // Show result section, hide problem detail
    problemDetail.style.display = 'none';
    submissionResult.style.display = 'block';
}

// Navigate back to problems list
function goBackToProblems() {
    // 清除题目详情内容
    problemTitle.textContent = '';
    problemDifficulty.textContent = '';
    problemLimits.textContent = '';
    problemDescriptionText.textContent = '';
    problemInput.textContent = '';
    problemOutput.textContent = '';
    examplesContainer.innerHTML = '';
    
    // 删除知识点标签
    const tagsContainer = document.querySelector('.problem-knowledge-tags');
    if (tagsContainer) {
        tagsContainer.remove();
    }
    
    // 显示题目列表，隐藏题目详情和提交结果
    problemsList.style.display = 'block';
    problemDetail.style.display = 'none';
    submissionResult.style.display = 'none';
    
    // 添加加载提示
    problemsContainer.innerHTML = '<div class="loading">刷新题目列表中...</div>';
    
    // 重新加载题目列表以获取最新状态
    loadProblems().then(() => {
        // 重新应用已有的过滤器
        applyFilters();
        
        // 如果有激活的过滤器，高亮对应选项卡
        document.querySelectorAll('.filter-tab').forEach(tab => {
            if (tab.getAttribute('data-filter') === activeFilter) {
                tab.classList.add('active');
            } else {
                tab.classList.remove('active');
            }
        });
    }).catch(error => {
        console.error('刷新题目列表失败:', error);
        problemsContainer.innerHTML = '<div class="error">刷新题目列表时出错，请稍后再试。</div>';
    });
}

// 打开导入对话框
function openImportDialog() {
    importDialog.style.display = 'block';
    importJsonTextarea.value = ''; // 清空文本框
}

// 关闭导入对话框
function closeImportDialog() {
    importDialog.style.display = 'none';
}

// 提交导入题目
async function submitImportProblems() {
    const jsonText = importJsonTextarea.value.trim();
    if (!jsonText) {
        alert('请输入JSON格式的题目数据');
        return;
    }
    
    try {
        // 验证JSON格式
        const data = JSON.parse(jsonText);
        
        // 验证基本结构
        if (!data.problems || !Array.isArray(data.problems) || data.problems.length === 0) {
            alert('JSON格式不正确，请检查problems数组');
            return;
        }
        
        // 禁用按钮防止重复提交
        submitImportBtn.disabled = true;
        submitImportBtn.textContent = '导入中...';
        
        // 提交请求
        const response = await fetch('/api/problems/import', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: jsonText
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            alert(`导入失败: ${errorData.error || '未知错误'}`);
            submitImportBtn.disabled = false;
            submitImportBtn.textContent = '导入';
            return;
        }
        
        const result = await response.json();
        alert(`成功导入${result.count}个题目`);
        
        // 关闭对话框并重新加载题目列表
        closeImportDialog();
        loadProblems();
        
    } catch (error) {
        console.error('JSON解析或导入错误:', error);
        alert(`导入错误: ${error.message || '请检查JSON格式'}`);
    } finally {
        // 恢复按钮状态
        submitImportBtn.disabled = false;
        submitImportBtn.textContent = '导入';
    }
}

// Helper functions
function getStatusClass(problem) {
    if (problem.solved) {
        return 'status-solved';
    } else if (problem.attempted) {
        return 'status-attempted';
    } else {
        return 'status-unsolved';
    }
}

function getStatusText(problem) {
    if (problem.solved) {
        return '已解决';
    } else if (problem.attempted) {
        return '尝试中';
    } else {
        return '未做';
    }
} 