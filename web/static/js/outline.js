// 大纲题库和知识点测试相关功能

// DOM 元素引用
const outlineSectionSelect = document.getElementById('outlineSectionSelect');
const questionCountInput = document.getElementById('questionCountInput');
const loadOutlineQuestionsBtn = document.getElementById('loadOutlineQuestionsBtn');
const generateQuestionsBtn = document.getElementById('generateQuestionsBtn');
const createQuizBtn = document.getElementById('createQuizBtn');
const outlineQuestions = document.getElementById('outlineQuestions');
const createNewQuizBtn = document.getElementById('createNewQuizBtn');
const quizzesList = document.getElementById('quizzesList');
const quizDetail = document.getElementById('quizDetail');
const quizTitle = document.getElementById('quizTitle');
const quizDescription = document.getElementById('quizDescription');
const quizQuestions = document.getElementById('quizQuestions');
const backToQuizzesBtn = document.getElementById('backToQuizzesBtn');
const submitQuizBtn = document.getElementById('submitQuizBtn');
const quizResult = document.getElementById('quizResult');
const quizScore = document.getElementById('quizScore');
const quizScoreBar = document.getElementById('quizScoreBar');
const quizResultDetails = document.getElementById('quizResultDetails');
const backToQuizzesFromResultBtn = document.getElementById('backToQuizzesFromResultBtn');
const saveQuizBtn = document.getElementById('saveQuizBtn');

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    // 事件监听器
    if (loadOutlineQuestionsBtn) {
        loadOutlineQuestionsBtn.addEventListener('click', loadOutlineQuestions);
    }
    
    if (generateQuestionsBtn) {
        generateQuestionsBtn.addEventListener('click', showGenerateQuestionsModal);
    }
    
    if (createQuizBtn) {
        createQuizBtn.addEventListener('click', showCreateQuizModal);
    }
    
    if (createNewQuizBtn) {
        createNewQuizBtn.addEventListener('click', showCreateQuizModal);
    }
    
    if (backToQuizzesBtn) {
        backToQuizzesBtn.addEventListener('click', backToQuizzes);
    }
    
    if (submitQuizBtn) {
        submitQuizBtn.addEventListener('click', submitQuiz);
    }
    
    if (backToQuizzesFromResultBtn) {
        backToQuizzesFromResultBtn.addEventListener('click', backToQuizzes);
    }
    
    if (saveQuizBtn) {
        saveQuizBtn.addEventListener('click', createQuiz);
    }
    
    // 加载测试列表
    document.querySelectorAll('#mainTabs button').forEach(button => {
        button.addEventListener('click', function(e) {
            if (e.target.id === 'quizzes-tab') {
                loadQuizzes();
            }
        });
    });
});

// 加载大纲题目
async function loadOutlineQuestions() {
    try {
        const section = outlineSectionSelect.value;
        const count = questionCountInput.value;
        
        if (!section) {
            alert('请选择一个大纲章节');
            return;
        }
        
        outlineQuestions.innerHTML = '<div class="text-center">加载中...</div>';
        
        const response = await fetch(`/api/outline/questions?section=${section}&count=${count}`);
        
        if (!response.ok) {
            throw new Error('加载题目失败');
        }
        
        const questions = await response.json();
        renderOutlineQuestions(questions);
    } catch (error) {
        console.error('Error loading outline questions:', error);
        outlineQuestions.innerHTML = `<div class="alert alert-danger">加载题目失败: ${error.message}</div>`;
    }
}

// 渲染大纲题目
function renderOutlineQuestions(questions) {
    if (!questions || questions.length === 0) {
        outlineQuestions.innerHTML = '<div class="alert alert-info">没有找到题目</div>';
        return;
    }
    
    outlineQuestions.innerHTML = '';
    
    // 按章节分组题目
    const questionsBySection = {};
    questions.forEach(question => {
        const sectionPrefix = question.outline_ref.split('.').slice(0, 2).join('.');
        if (!questionsBySection[sectionPrefix]) {
            questionsBySection[sectionPrefix] = [];
        }
        questionsBySection[sectionPrefix].push(question);
    });
    
    // 渲染每个章节的题目
    Object.keys(questionsBySection).forEach(section => {
        const sectionDiv = document.createElement('div');
        sectionDiv.classList.add('outline-section');
        
        // 章节标题
        const sectionTitle = document.createElement('h4');
        sectionTitle.textContent = `${section} 章节题目`;
        sectionDiv.appendChild(sectionTitle);
        
        // 章节题目
        questionsBySection[section].forEach((question, index) => {
            const questionCard = createQuestionCard(question, index + 1);
            sectionDiv.appendChild(questionCard);
        });
        
        outlineQuestions.appendChild(sectionDiv);
    });
}

// 创建题目卡片
function createQuestionCard(question, index) {
    const card = document.createElement('div');
    card.classList.add('card', 'mb-3');
    
    // 题目类型标识
    let typeText = '';
    switch (question.type) {
        case 'multiple_choice': typeText = '选择题'; break;
        case 'fill_blank': typeText = '填空题'; break;
        case 'short_answer': typeText = '简答题'; break;
        default: typeText = '未知类型';
    }
    
    // 题目难度显示
    const difficultyStars = '★'.repeat(Math.min(question.difficulty, 5)) + 
                           '☆'.repeat(Math.max(0, 5 - Math.min(question.difficulty, 5)));
    
    // 选择题选项
    let optionsHtml = '';
    if (question.type === 'multiple_choice' && question.options && question.options.length > 0) {
        optionsHtml = '<div class="mt-3">';
        question.options.forEach((option, i) => {
            const optionLetter = String.fromCharCode(65 + i); // A, B, C, D...
            optionsHtml += `
                <div class="question-option">
                    <input type="radio" id="q${question.id}_option${i}" name="q${question.id}" value="${optionLetter}">
                    <label for="q${question.id}_option${i}">${optionLetter}. ${option}</label>
                </div>
            `;
        });
        optionsHtml += '</div>';
    } else if (question.type === 'fill_blank') {
        // 填空题输入框
        optionsHtml = `
            <div class="mt-3">
                <input type="text" class="form-control" id="q${question.id}_answer" placeholder="请输入答案">
            </div>
        `;
    } else if (question.type === 'short_answer') {
        // 简答题文本框
        optionsHtml = `
            <div class="mt-3">
                <textarea class="form-control" id="q${question.id}_answer" rows="3" placeholder="请输入答案"></textarea>
            </div>
        `;
    }
    
    card.innerHTML = `
        <div class="card-header d-flex justify-content-between align-items-center">
            <span>${index}. [${typeText}] ${question.outline_ref}</span>
            <span class="badge bg-info">${difficultyStars}</span>
        </div>
        <div class="card-body">
            <p class="card-text">${question.content}</p>
            ${optionsHtml}
            <div class="question-explanation mt-3" style="display: none;" id="explanation_${question.id}">
                <p><strong>正确答案:</strong> ${question.answer}</p>
                <p><strong>解释:</strong> ${question.explanation}</p>
            </div>
            <button class="btn btn-sm btn-outline-secondary mt-2" onclick="toggleExplanation(${question.id})">查看解析</button>
        </div>
    `;
    
    return card;
}

// 切换显示题目解析
function toggleExplanation(questionId) {
    const explanation = document.getElementById(`explanation_${questionId}`);
    if (explanation.style.display === 'none') {
        explanation.style.display = 'block';
    } else {
        explanation.style.display = 'none';
    }
}

// 显示生成题目模态框
function showGenerateQuestionsModal() {
    // 使用Bootstrap的模态框显示生成题目的选项
    alert('此功能正在开发中，请使用"加载题目"按钮和下拉菜单进行测试');
}

// 显示创建测试模态框
function showCreateQuizModal() {
    // 使用Bootstrap的模态框
    const createQuizModal = new bootstrap.Modal(document.getElementById('createQuizModal'));
    createQuizModal.show();
}

// 创建测试
async function createQuiz() {
    try {
        const title = document.getElementById('quizTitleInput').value;
        const description = document.getElementById('quizDescriptionInput').value;
        const section = document.getElementById('quizSectionSelect').value;
        const count = document.getElementById('quizQuestionCountInput').value;
        
        if (!title) {
            alert('请输入测试标题');
            return;
        }
        
        if (!section) {
            alert('请选择大纲章节');
            return;
        }
        
        const response = await fetch('/api/quizzes', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                title,
                description,
                section,
                count: parseInt(count)
            })
        });
        
        if (!response.ok) {
            throw new Error('创建测试失败');
        }
        
        const quiz = await response.json();
        
        // 关闭模态框
        const createQuizModal = bootstrap.Modal.getInstance(document.getElementById('createQuizModal'));
        createQuizModal.hide();
        
        // 切换到测试列表并刷新
        document.getElementById('quizzes-tab').click();
        loadQuizzes();
        
    } catch (error) {
        console.error('Error creating quiz:', error);
        alert(`创建测试失败: ${error.message}`);
    }
}

// 加载测试列表
async function loadQuizzes() {
    try {
        quizzesList.innerHTML = '<div class="col-12 text-center">加载中...</div>';
        
        const response = await fetch('/api/quizzes');
        
        if (!response.ok) {
            throw new Error('加载测试失败');
        }
        
        const quizzes = await response.json();
        renderQuizzes(quizzes);
    } catch (error) {
        console.error('Error loading quizzes:', error);
        quizzesList.innerHTML = `<div class="col-12 alert alert-danger">加载测试失败: ${error.message}</div>`;
    }
}

// 渲染测试列表
function renderQuizzes(quizzes) {
    if (!quizzes || quizzes.length === 0) {
        quizzesList.innerHTML = '<div class="col-12 alert alert-info">没有找到测试</div>';
        return;
    }
    
    quizzesList.innerHTML = '';
    
    quizzes.forEach(quiz => {
        const col = document.createElement('div');
        col.classList.add('col-md-4', 'mb-4');
        
        const tagsBadges = quiz.knowledge_tag.map(tag => 
            `<span class="badge bg-secondary me-1">${tag}</span>`
        ).join('');
        
        col.innerHTML = `
            <div class="card quiz-card" data-quiz-id="${quiz.id}">
                <div class="card-body">
                    <h5 class="card-title">${quiz.title}</h5>
                    <p class="card-text">${quiz.description || '无描述'}</p>
                    <p class="small text-muted">题目数量: ${quiz.question_ids ? quiz.question_ids.length : 0}</p>
                    <div class="mb-2">${tagsBadges}</div>
                    <p class="small text-muted">创建时间: ${formatDateTime(quiz.created_at)}</p>
                    <button class="btn btn-primary btn-sm">开始测试</button>
                </div>
            </div>
        `;
        
        // 添加点击事件
        const startQuizBtn = col.querySelector('.btn');
        startQuizBtn.addEventListener('click', () => loadQuiz(quiz.id));
        
        quizzesList.appendChild(col);
    });
}

// 加载测试详情
async function loadQuiz(quizId) {
    try {
        quizDetail.style.display = 'none';
        quizResult.style.display = 'none';
        
        // 显示加载中
        quizzesList.innerHTML = '<div class="col-12 text-center">加载中...</div>';
        
        const response = await fetch(`/api/quizzes/${quizId}`);
        
        if (!response.ok) {
            throw new Error('加载测试失败');
        }
        
        const quiz = await response.json();
        
        // 隐藏测试列表，显示测试详情
        quizzesList.parentElement.style.display = 'none';
        quizDetail.style.display = 'block';
        
        // 更新标题和描述
        quizTitle.textContent = quiz.title;
        quizDescription.textContent = quiz.description || '无描述';
        
        // 渲染题目
        renderQuizQuestions(quiz.questions);
        
        // 保存当前测试ID，用于提交答案
        quizDetail.dataset.quizId = quizId;
        
    } catch (error) {
        console.error('Error loading quiz:', error);
        quizzesList.innerHTML = `<div class="col-12 alert alert-danger">加载测试失败: ${error.message}</div>`;
    }
}

// 渲染测试题目
function renderQuizQuestions(questions) {
    if (!questions || questions.length === 0) {
        quizQuestions.innerHTML = '<div class="alert alert-info">该测试没有题目</div>';
        return;
    }
    
    quizQuestions.innerHTML = '';
    
    questions.forEach((question, index) => {
        const questionDiv = document.createElement('div');
        questionDiv.classList.add('card', 'mb-3');
        questionDiv.dataset.questionId = question.id;
        
        // 题目类型标识
        let typeText = '';
        switch (question.type) {
            case 'multiple_choice': typeText = '选择题'; break;
            case 'fill_blank': typeText = '填空题'; break;
            case 'short_answer': typeText = '简答题'; break;
            default: typeText = '未知类型';
        }
        
        // 选择题选项
        let answerInputHtml = '';
        if (question.type === 'multiple_choice' && question.options && question.options.length > 0) {
            answerInputHtml = '<div class="mt-3">';
            question.options.forEach((option, i) => {
                const optionLetter = String.fromCharCode(65 + i); // A, B, C, D...
                answerInputHtml += `
                    <div class="question-option">
                        <input type="radio" id="q${question.id}_option${i}" name="q${question.id}" value="${optionLetter}" class="quiz-answer" data-question-id="${question.id}">
                        <label for="q${question.id}_option${i}">${optionLetter}. ${option}</label>
                    </div>
                `;
            });
            answerInputHtml += '</div>';
        } else if (question.type === 'fill_blank') {
            // 填空题输入框
            answerInputHtml = `
                <div class="mt-3">
                    <input type="text" class="form-control quiz-answer" id="q${question.id}_answer" 
                           placeholder="请输入答案" data-question-id="${question.id}">
                </div>
            `;
        } else if (question.type === 'short_answer') {
            // 简答题文本框
            answerInputHtml = `
                <div class="mt-3">
                    <textarea class="form-control quiz-answer" id="q${question.id}_answer" rows="3" 
                             placeholder="请输入答案" data-question-id="${question.id}"></textarea>
                </div>
            `;
        }
        
        questionDiv.innerHTML = `
            <div class="card-header">
                ${index + 1}. [${typeText}] ${question.outline_ref || ''}
            </div>
            <div class="card-body">
                <p class="card-text">${question.content}</p>
                ${answerInputHtml}
            </div>
        `;
        
        quizQuestions.appendChild(questionDiv);
    });
}

// 提交测试
async function submitQuiz() {
    try {
        const quizId = quizDetail.dataset.quizId;
        
        if (!quizId) {
            alert('测试ID无效');
            return;
        }
        
        // 收集所有答案
        const answers = [];
        document.querySelectorAll('.quiz-answer').forEach(input => {
            const questionId = input.dataset.questionId;
            let answer = '';
            
            if (input.type === 'radio') {
                // 选择题
                if (input.checked) {
                    answer = input.value;
                } else {
                    return; // 跳过未选中的选项
                }
            } else if (input.type === 'text' || input.tagName.toLowerCase() === 'textarea') {
                // 填空题或简答题
                answer = input.value;
            }
            
            // 只添加有答案的题目
            if (answer.trim()) {
                answers.push({
                    question_id: parseInt(questionId),
                    answer: answer.trim()
                });
            }
        });
        
        if (answers.length === 0) {
            alert('请至少回答一道题目');
            return;
        }
        
        // 发送答案
        const response = await fetch(`/api/quizzes/${quizId}/submit`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                user_id: currentUser.id, // 从全局变量获取
                answers: answers
            })
        });
        
        if (!response.ok) {
            throw new Error('提交答案失败');
        }
        
        const result = await response.json();
        
        // 显示结果
        showQuizResult(result);
        
    } catch (error) {
        console.error('Error submitting quiz:', error);
        alert(`提交测试失败: ${error.message}`);
    }
}

// 显示测试结果
function showQuizResult(result) {
    // 隐藏测试详情，显示结果
    quizDetail.style.display = 'none';
    quizResult.style.display = 'block';
    
    // 更新分数
    const scorePercent = result.score * 100;
    quizScore.textContent = `得分: ${scorePercent.toFixed(1)}%`;
    quizScoreBar.style.width = `${scorePercent}%`;
    quizScoreBar.textContent = `${scorePercent.toFixed(1)}%`;
    quizScoreBar.setAttribute('aria-valuenow', scorePercent);
    
    // 根据分数设置进度条颜色
    if (scorePercent >= 80) {
        quizScoreBar.classList.remove('bg-warning', 'bg-danger');
        quizScoreBar.classList.add('bg-success');
    } else if (scorePercent >= 60) {
        quizScoreBar.classList.remove('bg-success', 'bg-danger');
        quizScoreBar.classList.add('bg-warning');
    } else {
        quizScoreBar.classList.remove('bg-success', 'bg-warning');
        quizScoreBar.classList.add('bg-danger');
    }
    
    // 渲染答题详情
    renderQuizResultDetails(result.details);
}

// 渲染测试结果详情
function renderQuizResultDetails(details) {
    if (!details || details.length === 0) {
        quizResultDetails.innerHTML = '<div class="alert alert-info">没有详细结果</div>';
        return;
    }
    
    quizResultDetails.innerHTML = '';
    
    details.forEach((detail, index) => {
        const detailDiv = document.createElement('div');
        detailDiv.classList.add('card', 'mb-3');
        
        // 设置正确/错误样式
        const cardClass = detail.is_correct ? 'border-success' : 'border-danger';
        detailDiv.classList.add(cardClass);
        
        // 题目类型标识
        let typeText = '';
        switch (detail.question.type) {
            case 'multiple_choice': typeText = '选择题'; break;
            case 'fill_blank': typeText = '填空题'; break;
            case 'short_answer': typeText = '简答题'; break;
            default: typeText = '未知类型';
        }
        
        detailDiv.innerHTML = `
            <div class="card-header d-flex justify-content-between align-items-center">
                <span>${index + 1}. [${typeText}] ${detail.question.outline_ref || ''}</span>
                <span class="badge ${detail.is_correct ? 'bg-success' : 'bg-danger'}">
                    ${detail.is_correct ? '正确' : '错误'}
                </span>
            </div>
            <div class="card-body">
                <p class="card-text">${detail.question.content}</p>
                <div class="mt-3">
                    <p><strong>你的答案:</strong> ${detail.user_answer}</p>
                    <p><strong>正确答案:</strong> ${detail.question.answer}</p>
                </div>
                <div class="question-explanation mt-3">
                    <p><strong>解释:</strong> ${detail.question.explanation}</p>
                </div>
            </div>
        `;
        
        quizResultDetails.appendChild(detailDiv);
    });
}

// 返回测试列表
function backToQuizzes() {
    // 隐藏测试详情和结果，显示列表
    quizDetail.style.display = 'none';
    quizResult.style.display = 'none';
    quizzesList.parentElement.style.display = 'block';
    
    // 重新加载测试列表
    loadQuizzes();
}

// 将全局函数暴露到window对象
window.toggleExplanation = toggleExplanation; 