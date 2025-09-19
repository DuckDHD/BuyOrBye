/**
 * Enhanced ChatInterface - Mobile-first with smooth animations
 * Provides instant, intuitive chat experience for purchase decisions
 */
class ChatInterface {
    constructor() {
        this.sessionId = this.getOrCreateSessionId();
        this.conversationCount = 0;
        this.decisionCount = 0;
        this.isProcessing = false;

        // DOM elements
        this.messagesContainer = document.getElementById('chat-messages');
        this.userInput = document.getElementById('user-input');
        this.sendBtn = document.getElementById('send-btn');
        this.chatForm = document.getElementById('chat-form');
        this.charCount = document.getElementById('char-count');

        this.initializeEventListeners();
        this.loadSessionHistory();
        this.setupQuickStarters();
    }

    /**
     * Get existing session ID from localStorage or create new one
     */
    getOrCreateSessionId() {
        // Try to get existing session from localStorage
        let sessionId = localStorage.getItem('buyorbye_session');
        if (!sessionId) {
            sessionId = 'anon_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now().toString(36);
            localStorage.setItem('buyorbye_session', sessionId);
            localStorage.setItem('buyorbye_session_created', Date.now().toString());
        }
        return sessionId;
    }

    /**
     * Load previous session history if available
     */
    async loadSessionHistory() {
        try {
            const response = await fetch(`/api/chat/history?session_id=${this.sessionId}`);
            if (response.ok) {
                const data = await response.json();
                this.displaySessionHistory(data);
                this.checkAccountPrompt(data);
            }
        } catch (error) {
            console.log('No previous session found or session expired');
            // Start fresh - this is normal for new sessions
        }
    }

    /**
     * Display previous session history
     */
    displaySessionHistory(data) {
        // Clear welcome message if history exists
        if (data.messages && data.messages.length > 0) {
            this.messagesContainer.innerHTML = '';

            // Display previous messages
            data.messages.forEach(msg => {
                this.addMessage(msg.content, msg.role, false, false);
            });

            // Show decision summaries
            if (data.decisions && data.decisions.length > 0) {
                this.showSessionSummary(data.decisions);
            }

            this.scrollToBottom();
        }
    }

    /**
     * Show session summary with previous decisions
     */
    showSessionSummary(decisions) {
        const summaryHtml = `
            <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 my-4">
                <h4 class="font-medium text-blue-900 dark:text-blue-100 mb-2">📋 Your Previous Decisions</h4>
                <div class="space-y-2">
                    ${decisions.slice(0, 3).map(decision => `
                        <div class="text-sm">
                            <span class="font-medium">${decision.item_name || 'Purchase'}</span>
                            ${decision.price > 0 ? `($${decision.price})` : ''} -
                            <span class="px-2 py-1 rounded text-xs ${this.getDecisionBadgeClass(decision.decision)}">
                                ${decision.decision.toUpperCase()}
                            </span>
                        </div>
                    `).join('')}
                    ${decisions.length > 3 ? `<div class="text-xs text-blue-600 dark:text-blue-400">...and ${decisions.length - 3} more</div>` : ''}
                </div>
            </div>
        `;

        this.messagesContainer.insertAdjacentHTML('beforeend', summaryHtml);
    }

    /**
     * Get CSS classes for decision badges
     */
    getDecisionBadgeClass(decision) {
        switch (decision) {
            case 'buy': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
            case 'wait': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
            case 'bye': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
            default: return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
        }
    }

    /**
     * Check if we should prompt for account creation
     */
    checkAccountPrompt(data) {
        if (data.should_prompt_signup && !localStorage.getItem('buyorbye_account_prompted')) {
            setTimeout(() => {
                this.showAccountPrompt(data.stats);
            }, 2000); // Show after 2 seconds
        }
    }

    /**
     * Initialize all event listeners
     */
    initializeEventListeners() {
        // Main form submission
        this.chatForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.sendMessage();
        });

        // Auto-resize input and character count
        this.userInput.addEventListener('input', this.handleInputChange.bind(this));

        // Enter key handling
        this.userInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Account modal events
        const accountBtn = document.getElementById('account-btn');
        if (accountBtn) {
            accountBtn.addEventListener('click', this.showAccountModal.bind(this));
        }

        const accountForm = document.getElementById('account-form');
        if (accountForm) {
            accountForm.addEventListener('submit', this.createAccount.bind(this));
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') this.closeAccountModal();
        });

        // Touch-friendly interactions for mobile
        this.setupTouchInteractions();
    }

    /**
     * Setup quick starter buttons
     */
    setupQuickStarters() {
        document.querySelectorAll('.quick-starter').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const text = e.target.textContent.replace(/"/g, '');
                this.userInput.value = text;
                this.sendMessage();
            });
        });
    }

    /**
     * Handle input changes with auto-resize and character count
     */
    handleInputChange(e) {
        const input = e.target;
        const charCount = this.charCount;

        // Show character count for longer messages
        if (input.value.length > 100) {
            charCount.textContent = `${input.value.length}/500`;
            charCount.classList.remove('hidden');
        } else {
            charCount.classList.add('hidden');
        }

        // Auto-expand for longer messages on desktop
        if (window.innerWidth > 640 && input.value.length > 50) {
            input.style.height = 'auto';
            input.style.height = Math.min(input.scrollHeight, 120) + 'px';
        }
    }

    /**
     * Setup touch interactions for mobile
     */
    setupTouchInteractions() {
        // Prevent zoom on iOS when focusing input
        if (/iPad|iPhone|iPod/.test(navigator.userAgent)) {
            this.userInput.addEventListener('focus', (e) => {
                e.target.style.fontSize = '16px';
            });
        }

        // Add haptic feedback for modern mobile browsers
        this.sendBtn.addEventListener('click', () => {
            if ('vibrate' in navigator) {
                navigator.vibrate(50);
            }
        });
    }

    /**
     * Update character count display
     */
    updateCharCount() {
        const currentLength = this.userInput.value.length;
        this.charCount.textContent = `${currentLength}/500`;

        if (currentLength > 450) {
            this.charCount.classList.add('text-red-500');
            this.charCount.classList.remove('text-gray-400');
        } else {
            this.charCount.classList.add('text-gray-400');
            this.charCount.classList.remove('text-red-500');
        }
    }

    /**
     * Send a message to the AI with enhanced animations
     */
    async sendMessage() {
        const message = this.userInput.value.trim();

        if (!message || this.isProcessing || message.length > 500) {
            return;
        }

        // UI updates
        this.addMessage(message, 'user');
        this.userInput.value = '';
        this.userInput.style.height = 'auto';
        this.charCount.classList.add('hidden');
        this.setLoading(true);
        this.conversationCount++;

        try {
            const response = await fetch('/api/chat', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    message: message,
                    session_id: this.sessionId
                })
            });

            if (!response.ok) throw new Error('Network error');

            const data = await response.json();
            this.handleResponse(data);

        } catch (error) {
            console.error('Chat error:', error);
            this.addMessage('Sorry, something went wrong. Please try again.', 'ai');
        } finally {
            this.setLoading(false);
        }
    }

    /**
     * Handle different response types
     */
    handleResponse(data) {
        if (data.status === 'decision') {
            this.showDecision(data.decision);
            this.decisionCount++;

            // Show account prompt after 2 decisions
            if (this.decisionCount >= 2 && !this.hasAccount()) {
                setTimeout(() => this.showAccountPrompt(), 2000);
            }
        } else if (data.status === 'need_info') {
            this.addMessage(data.response, 'ai');
            if (data.questions) {
                this.showQuickQuestions(data.questions);
            }
        } else {
            this.addMessage(data.response, 'ai');
        }
    }

    /**
     * Add a message with smooth animations and mobile-optimized styling
     */
    addMessage(text, sender, isError = false) {
        const messageDiv = document.createElement('div');
        messageDiv.className = 'flex items-start space-x-3';

        const avatar = sender === 'user'
            ? '<div class="w-8 h-8 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center flex-shrink-0"><span class="text-gray-600 dark:text-gray-300 text-sm">👤</span></div>'
            : '<div class="w-8 h-8 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center flex-shrink-0"><span class="text-blue-600 dark:text-blue-400 text-sm">🤖</span></div>';

        const messageClass = sender === 'user'
            ? 'bg-blue-600 text-white p-3 rounded-lg rounded-tr-none max-w-xs sm:max-w-sm ml-auto'
            : isError
                ? 'bg-red-50 dark:bg-red-900/20 text-red-800 dark:text-red-200 p-3 rounded-lg rounded-tl-none max-w-xs sm:max-w-sm border border-red-200 dark:border-red-800'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200 p-3 rounded-lg rounded-tl-none max-w-xs sm:max-w-sm';

        messageDiv.innerHTML = `
            ${sender === 'user' ? '' : avatar}
            <div class="${messageClass}">
                ${this.formatMessage(text)}
            </div>
            ${sender === 'user' ? avatar : ''}
        `;

        if (sender === 'user') {
            messageDiv.classList.add('flex-row-reverse');
        }

        this.messagesContainer.appendChild(messageDiv);
        this.scrollToBottom();

        // Animate message appearance
        messageDiv.style.opacity = '0';
        messageDiv.style.transform = 'translateY(10px)';
        requestAnimationFrame(() => {
            messageDiv.style.transition = 'all 0.3s ease';
            messageDiv.style.opacity = '1';
            messageDiv.style.transform = 'translateY(0)';
        });
    }

    /**
     * Format message text with basic markup support
     */
    formatMessage(text) {
        return text
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            .replace(/\*(.*?)\*/g, '<em>$1</em>')
            .replace(/(https?:\/\/[^\s]+)/g, '<a href="$1" target="_blank" class="text-blue-600 underline">$1</a>');
    }

    /**
     * Show account creation prompt
     */
    showAccountPrompt(stats) {
        const promptHtml = `
            <div id="account-prompt" class="bg-gradient-to-r from-green-50 to-blue-50 dark:from-green-900/20 dark:to-blue-900/20 border border-green-200 dark:border-green-800 rounded-xl p-6 my-4 shadow-sm">
                <div class="flex items-start space-x-4">
                    <div class="bg-green-600 text-white rounded-full p-2 flex-shrink-0">
                        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                        </svg>
                    </div>
                    <div class="flex-1">
                        <h4 class="font-bold text-green-900 dark:text-green-100 text-lg mb-2">💾 Save Your Decision History</h4>
                        <p class="text-green-800 dark:text-green-200 mb-3">
                            You've made ${stats?.decision_count || 'several'} smart decisions with our AI!
                            Create a free account to save your progress and get even more personalized recommendations.
                        </p>
                        <div class="text-sm text-green-700 dark:text-green-300 mb-4">
                            ✨ Your conversation history will be preserved<br>
                            📊 Get personalized spending insights<br>
                            🎯 Receive smarter recommendations over time
                        </div>
                        <div class="flex space-x-3">
                            <button onclick="window.chatInterface.showAccountForm()"
                                class="bg-green-600 hover:bg-green-700 text-white px-6 py-2 rounded-lg font-medium transition-colors">
                                Create Free Account
                            </button>
                            <button onclick="window.chatInterface.dismissAccountPrompt()"
                                class="bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 px-4 py-2 rounded-lg transition-colors">
                                Maybe Later
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.messagesContainer.insertAdjacentHTML('beforeend', promptHtml);
        this.scrollToBottom();
    }

    /**
     * Show account creation form
     */
    showAccountForm() {
        const formHtml = `
            <div id="account-form" class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl p-6 my-4 shadow-lg">
                <h4 class="font-bold text-gray-900 dark:text-gray-100 text-lg mb-4">Create Your Account</h4>
                <form id="signup-form" class="space-y-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label>
                        <input type="text" id="signup-name" required
                            class="w-full p-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
                        <input type="email" id="signup-email" required
                            class="w-full p-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Password</label>
                        <input type="password" id="signup-password" required minlength="8"
                            class="w-full p-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Minimum 8 characters</div>
                    </div>
                    <div class="flex space-x-3 pt-2">
                        <button type="submit" class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg font-medium transition-colors">
                            Create Account
                        </button>
                        <button type="button" onclick="window.chatInterface.hideAccountForm()"
                            class="bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 px-4 py-2 rounded-lg transition-colors">
                            Cancel
                        </button>
                    </div>
                </form>
            </div>
        `;

        // Replace the prompt with the form
        const prompt = document.getElementById('account-prompt');
        if (prompt) {
            prompt.outerHTML = formHtml;
        } else {
            this.messagesContainer.insertAdjacentHTML('beforeend', formHtml);
        }

        // Add form submission handler
        document.getElementById('signup-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.createAccount();
        });

        this.scrollToBottom();
    }

    /**
     * Create account from session
     */
    async createAccount() {
        const name = document.getElementById('signup-name').value.trim();
        const email = document.getElementById('signup-email').value.trim();
        const password = document.getElementById('signup-password').value;

        if (!name || !email || !password) {
            alert('Please fill in all fields');
            return;
        }

        try {
            const response = await fetch('/api/chat/create-account', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    session_id: this.sessionId,
                    name: name,
                    email: email,
                    password: password
                })
            });

            const data = await response.json();

            if (response.ok) {
                // Hide the form
                this.hideAccountForm();

                // Show success message
                this.addMessage(`🎉 Account created successfully! ${data.message}`, 'ai');

                // Clear session from localStorage since it's now associated with an account
                localStorage.removeItem('buyorbye_session');
                localStorage.setItem('buyorbye_account_created', 'true');

                // Optionally redirect to login or dashboard
                if (data.requires_login) {
                    setTimeout(() => {
                        window.location.href = '/auth/login';
                    }, 2000);
                }
            } else {
                alert(data.error || 'Failed to create account');
            }
        } catch (error) {
            console.error('Account creation error:', error);
            alert('Failed to create account. Please try again.');
        }
    }

    /**
     * Dismiss account prompt
     */
    dismissAccountPrompt() {
        const prompt = document.getElementById('account-prompt');
        if (prompt) {
            prompt.remove();
        }
        localStorage.setItem('buyorbye_account_prompted', 'true');
    }

    /**
     * Hide account form
     */
    hideAccountForm() {
        const form = document.getElementById('account-form');
        if (form) {
            form.remove();
        }
    }

    /**
     * Get CSS classes for message styling
     */
    getMessageClasses(sender, isError) {
        const baseClasses = 'max-w-2xl p-4 rounded-lg border';

        if (sender === 'user') {
            return `${baseClasses} bg-blue-600 text-white ml-12`;
        } else if (isError) {
            return `${baseClasses} bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 mr-12`;
        } else {
            return `${baseClasses} bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 mr-12`;
        }
    }

    /**
     * Set loading state for the interface
     */
    setLoadingState(isLoading) {
        this.submitBtn.disabled = isLoading;
        this.userInput.disabled = isLoading;

        if (isLoading) {
            this.submitBtn.innerHTML = `
                <span>Processing...</span>
                <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
            `;
            this.aiThinking.classList.remove('hidden');
        } else {
            this.submitBtn.innerHTML = `
                <span>Ask AI</span>
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"></path>
                </svg>
            `;
            this.aiThinking.classList.add('hidden');
        }
    }

    /**
     * Show an error message
     */
    showError(message) {
        this.addMessage(message, 'ai', true);
    }

    /**
     * Show enhanced decision with mobile-optimized styling
     */
    showDecision(decision) {
        const decisionDiv = document.createElement('div');
        decisionDiv.className = 'decision-card bg-gradient-to-r from-green-50 to-blue-50 dark:from-green-900/20 dark:to-blue-900/20 p-4 rounded-lg border-l-4 border-green-500 mx-0 sm:mx-11 mt-4';

        const emoji = decision.recommendation === 'buy' ? '✅' : decision.recommendation === 'wait' ? '⏰' : '❌';
        const color = decision.recommendation === 'buy' ? 'text-green-600' : decision.recommendation === 'wait' ? 'text-yellow-600' : 'text-red-600';

        decisionDiv.innerHTML = `
            <div class="flex items-center gap-2 mb-2">
                <span class="text-2xl">${emoji}</span>
                <h3 class="font-bold text-lg ${color}">${decision.recommendation.toUpperCase()}</h3>
                <span class="text-sm text-gray-500 ml-auto">Confidence: ${Math.round(decision.confidence * 100)}%</span>
            </div>
            <p class="text-gray-700 dark:text-gray-300 leading-relaxed">${decision.reasoning}</p>
            ${decision.wait_period ? `<p class="text-sm text-gray-600 dark:text-gray-400 mt-2 flex items-center"><span class="mr-1">💡</span> Suggested wait: ${decision.wait_period} days</p>` : ''}
        `;

        this.messagesContainer.appendChild(decisionDiv);
        this.scrollToBottom();

        // Animate decision card
        decisionDiv.style.opacity = '0';
        decisionDiv.style.transform = 'scale(0.95)';
        requestAnimationFrame(() => {
            decisionDiv.style.transition = 'all 0.4s ease';
            decisionDiv.style.opacity = '1';
            decisionDiv.style.transform = 'scale(1)';
        });
    }

    /**
     * Set loading state with typing indicator
     */
    setLoading(loading) {
        const sendBtn = this.sendBtn;
        const sendText = sendBtn.querySelector('.send-text');
        const sendLoading = sendBtn.querySelector('.send-loading');

        if (loading) {
            sendText.classList.add('hidden');
            sendLoading.classList.remove('hidden');
            sendBtn.disabled = true;
            this.showTypingIndicator();
        } else {
            sendText.classList.remove('hidden');
            sendLoading.classList.add('hidden');
            sendBtn.disabled = false;
            this.hideTypingIndicator();
        }
    }

    /**
     * Show typing indicator with animated dots
     */
    showTypingIndicator() {
        const indicator = document.createElement('div');
        indicator.className = 'typing-indicator flex items-start space-x-3';
        indicator.innerHTML = `
            <div class="w-8 h-8 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center flex-shrink-0">
                <span class="text-blue-600 dark:text-blue-400 text-sm">🤖</span>
            </div>
            <div class="bg-gray-100 dark:bg-gray-700 p-3 rounded-lg rounded-tl-none">
                <div class="flex space-x-1">
                    <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                    <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0.1s"></div>
                    <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
                </div>
            </div>
        `;

        this.messagesContainer.appendChild(indicator);
        this.scrollToBottom();
    }

    /**
     * Hide typing indicator
     */
    hideTypingIndicator() {
        const indicator = document.querySelector('.typing-indicator');
        if (indicator) indicator.remove();
    }

    /**
     * Show quick question buttons
     */
    showQuickQuestions(questions) {
        const questionsHtml = `
            <div class="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg border border-gray-200 dark:border-gray-700">
                <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">Quick answers:</p>
                <div class="flex flex-wrap gap-2">
                    ${questions.map(q => `
                        <button class="quick-question bg-blue-100 hover:bg-blue-200 dark:bg-blue-900 dark:hover:bg-blue-800 text-blue-800 dark:text-blue-200 px-3 py-2 rounded-full text-sm transition-colors"
                                onclick="window.chatInterface.selectQuestion('${q.replace(/'/g, "\\'")}')">
                            ${q}
                        </button>
                    `).join('')}
                </div>
            </div>
        `;

        const messageWrapper = document.createElement('div');
        messageWrapper.className = 'flex justify-start mb-4';
        messageWrapper.innerHTML = questionsHtml;

        this.messagesContainer.appendChild(messageWrapper);
        this.scrollToBottom();
    }

    /**
     * Handle quick question selection
     */
    selectQuestion(question) {
        this.userInput.value = question;
        this.updateCharCount();
        this.sendMessage();
    }

    /**
     * Ask a follow-up question
     */
    askFollowUp(question) {
        this.userInput.value = question;
        this.updateCharCount();
        this.sendMessage();
    }

    /**
     * Start a new decision process
     */
    startNew() {
        this.userInput.value = '';
        this.updateCharCount();
        this.userInput.focus();
        this.addMessage("What's your next purchase decision? I'm here to help!", 'ai');
    }

    /**
     * Show account prompt after user engagement
     */
    showAccountPrompt() {
        if (this.hasAccount()) return;

        const prompt = document.createElement('div');
        prompt.className = 'account-prompt bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 p-4 rounded-lg mx-0 sm:mx-11 mt-4';
        prompt.innerHTML = `
            <div class="flex items-start gap-3">
                <span class="text-2xl">💾</span>
                <div class="flex-1">
                    <h4 class="font-medium text-blue-900 dark:text-blue-100">Save Your Decisions</h4>
                    <p class="text-blue-700 dark:text-blue-300 text-sm mt-1">
                        You've made ${this.decisionCount} smart decisions! Create a free account to keep your history.
                    </p>
                    <div class="mt-3 flex gap-2">
                        <button onclick="chatInterface.showAccountModal()" class="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700 transition-colors">
                            Create Account
                        </button>
                        <button onclick="this.parentElement.parentElement.parentElement.parentElement.remove()" class="text-blue-600 text-sm hover:text-blue-800 transition-colors">
                            Maybe Later
                        </button>
                    </div>
                </div>
            </div>
        `;

        this.messagesContainer.appendChild(prompt);
        this.scrollToBottom();
    }

    /**
     * Show account creation modal
     */
    showAccountModal() {
        const modal = document.getElementById('account-modal');
        modal.classList.remove('hidden');
        modal.classList.add('flex');
        document.getElementById('account-name').focus();
    }

    /**
     * Close account creation modal
     */
    closeAccountModal() {
        const modal = document.getElementById('account-modal');
        modal.classList.add('hidden');
        modal.classList.remove('flex');
    }

    /**
     * Create account from session data
     */
    async createAccount(e) {
        e.preventDefault();

        const name = document.getElementById('account-name').value.trim();
        const email = document.getElementById('account-email').value.trim();
        const password = document.getElementById('account-password').value;

        if (!name || !email || !password) {
            return;
        }

        try {
            const response = await fetch('/api/chat/create-account', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    session_id: this.sessionId,
                    name: name,
                    email: email,
                    password: password
                })
            });

            const data = await response.json();

            if (response.ok) {
                this.closeAccountModal();
                this.addMessage(`🎉 Account created successfully! ${data.message}`, 'ai');

                // Update UI to show signed in state
                localStorage.setItem('buyorbye_user', JSON.stringify({ name, email }));
                document.getElementById('account-btn').textContent = `Hi, ${name.split(' ')[0]}!`;

                // Clear session from localStorage since it's now associated with an account
                localStorage.removeItem('buyorbye_session');
            } else {
                alert(data.error || 'Failed to create account');
            }
        } catch (error) {
            console.error('Account creation error:', error);
            alert('Failed to create account. Please try again.');
        }
    }

    /**
     * Check if user has an account
     */
    hasAccount() {
        return localStorage.getItem('buyorbye_user') !== null;
    }

    /**
     * Scroll chat to bottom with smooth behavior
     */
    scrollToBottom() {
        this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;

        // Smooth scroll on modern browsers
        if (this.messagesContainer.scrollTo) {
            this.messagesContainer.scrollTo({
                top: this.messagesContainer.scrollHeight,
                behavior: 'smooth'
            });
        }
    }
}

// Initialize chat interface when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Only initialize if we're on the chat page
    if (document.getElementById('chat-form')) {
        window.chatInterface = new ChatInterface();
    }
});

// Export for potential use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ChatInterface;
}