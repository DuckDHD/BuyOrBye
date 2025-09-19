/**
 * Structured Chat Flow - Enhanced Frontend for Step-by-Step Purchase Decisions
 * Handles multi-step data collection and AI-powered recommendation flow
 */

class StructuredChatFlow {
    constructor() {
        this.sessionId = this.generateSessionId();
        this.currentStep = 1;
        this.userData = {};
        this.isProcessing = false;

        this.init();
    }

    init() {
        this.setupEventListeners();
        this.initializeFirstStep();
    }

    generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    setupEventListeners() {
        // Form submission handler
        document.addEventListener('submit', (e) => {
            if (e.target.classList.contains('structured-chat-form')) {
                e.preventDefault();
                this.handleFormSubmission(e.target);
            }
        });

        // Dynamic form field handlers
        document.addEventListener('input', (e) => {
            if (e.target.classList.contains('structured-input')) {
                this.validateField(e.target);
            }
        });

        // Step navigation
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('step-nav-btn')) {
                e.preventDefault();
                const targetStep = parseInt(e.target.dataset.step);
                this.navigateToStep(targetStep);
            }
        });

        // Restart flow
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('restart-flow-btn')) {
                e.preventDefault();
                this.restartFlow();
            }
        });
    }

    initializeFirstStep() {
        this.showStep1();
    }

    async handleFormSubmission(form) {
        if (this.isProcessing) return;

        this.isProcessing = true;
        this.showLoadingState(form);

        try {
            const formData = new FormData(form);
            const response = this.buildStepResponse(formData);

            // Send to backend
            const result = await this.sendStepData(this.currentStep, response);

            if (result.error) {
                this.showError(result.error);
                return;
            }

            // Process successful response
            this.processStepResponse(result);

        } catch (error) {
            console.error('Form submission error:', error);
            this.showError('Something went wrong. Please try again.');
        } finally {
            this.isProcessing = false;
            this.hideLoadingState(form);
        }
    }

    buildStepResponse(formData) {
        const response = {};

        for (let [key, value] of formData.entries()) {
            response[key] = value;
        }

        return Object.values(response).join(' | ');
    }

    async sendStepData(step, response) {
        const payload = {
            session_id: this.sessionId,
            step: step,
            response: response
        };

        const csrfToken = document.querySelector('meta[name="csrf-token"]')?.content;

        const fetchResponse = await fetch('/api/structured-chat/step', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken || ''
            },
            body: JSON.stringify(payload)
        });

        if (!fetchResponse.ok) {
            throw new Error(`HTTP error! status: ${fetchResponse.status}`);
        }

        return await fetchResponse.json();
    }

    processStepResponse(result) {
        // Update current step
        this.currentStep = result.step;

        // Store session ID if provided
        if (result.session_id) {
            this.sessionId = result.session_id;
        }

        // Navigate to next step or show result
        if (result.step === 'result') {
            this.showRecommendation(result);
        } else {
            this.showNextStep(result);
        }
    }

    showStep1() {
        const container = document.getElementById('structured-chat-container');
        container.innerHTML = `
            <div class="step-container" data-step="1">
                <div class="step-header">
                    <h3 class="text-xl font-semibold text-gray-800">Step 1: What are you considering buying?</h3>
                    <div class="step-progress">
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: 20%"></div>
                        </div>
                        <span class="progress-text">Step 1 of 5</span>
                    </div>
                </div>

                <form class="structured-chat-form mt-6" data-step="1">
                    <div class="form-group">
                        <label for="item-description" class="block text-sm font-medium text-gray-700 mb-2">
                            Describe what you want to buy and why
                        </label>
                        <textarea
                            id="item-description"
                            name="item_description"
                            class="structured-input w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            rows="4"
                            placeholder="e.g., iPhone 15 Pro for $999 because my current phone is broken and affecting my work"
                            required></textarea>
                        <div class="field-error text-red-500 text-sm mt-1 hidden"></div>
                    </div>

                    <div class="form-actions mt-6">
                        <button type="submit" class="btn-primary w-full py-3 px-6 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                            Continue to Need vs Want Analysis
                        </button>
                    </div>
                </form>
            </div>
        `;
    }

    showNextStep(result) {
        const container = document.getElementById('structured-chat-container');

        switch (result.step) {
            case 2:
                this.showStep2(result);
                break;
            case 3:
                this.showStep3(result);
                break;
            case 4:
                this.showStep4(result);
                break;
            case 5:
                this.showStep5(result);
                break;
            default:
                this.showError('Invalid step received');
        }
    }

    showStep2(result) {
        const container = document.getElementById('structured-chat-container');
        container.innerHTML = `
            <div class="step-container" data-step="2">
                <div class="step-header">
                    <h3 class="text-xl font-semibold text-gray-800">${result.question}</h3>
                    <div class="step-progress">
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: 40%"></div>
                        </div>
                        <span class="progress-text">Step 2 of 5</span>
                    </div>
                </div>

                <form class="structured-chat-form mt-6" data-step="2">
                    <div class="form-group">
                        <div class="radio-group">
                            ${result.options.map((option, index) => `
                                <label class="radio-option flex items-center p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer">
                                    <input type="radio" name="need_want" value="${option}" class="mr-3" required>
                                    <span class="text-gray-700">${option}</span>
                                </label>
                            `).join('')}
                        </div>
                    </div>

                    <div class="form-group mt-4">
                        <label for="reasoning" class="block text-sm font-medium text-gray-700 mb-2">
                            ${result.follow_up}
                        </label>
                        <textarea
                            id="reasoning"
                            name="reasoning"
                            class="structured-input w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            rows="3"
                            placeholder="Explain your reasoning..."
                            required></textarea>
                    </div>

                    <div class="form-actions mt-6 flex gap-3">
                        <button type="button" class="step-nav-btn btn-secondary px-6 py-3 border border-gray-300 rounded-lg hover:bg-gray-50" data-step="1">
                            Back
                        </button>
                        <button type="submit" class="btn-primary flex-1 py-3 px-6 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                            Continue to Financial Information
                        </button>
                    </div>
                </form>
            </div>
        `;
    }

    showStep3(result) {
        const container = document.getElementById('structured-chat-container');
        container.innerHTML = `
            <div class="step-container" data-step="3">
                <div class="step-header">
                    <h3 class="text-xl font-semibold text-gray-800">${result.question}</h3>
                    <div class="step-progress">
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: 60%"></div>
                        </div>
                        <span class="progress-text">Step 3 of 5</span>
                    </div>
                </div>

                <form class="structured-chat-form mt-6" data-step="3">
                    <div class="grid md:grid-cols-2 gap-4">
                        ${result.fields.map((field, index) => `
                            <div class="form-group">
                                <label for="field-${index}" class="block text-sm font-medium text-gray-700 mb-2">
                                    ${field.label}
                                </label>
                                <input
                                    type="number"
                                    id="field-${index}"
                                    name="financial_${index}"
                                    class="structured-input w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    placeholder="${field.placeholder}"
                                    min="0"
                                    step="0.01"
                                    required>
                                <div class="field-error text-red-500 text-sm mt-1 hidden"></div>
                            </div>
                        `).join('')}
                    </div>

                    <div class="form-actions mt-6 flex gap-3">
                        <button type="button" class="step-nav-btn btn-secondary px-6 py-3 border border-gray-300 rounded-lg hover:bg-gray-50" data-step="2">
                            Back
                        </button>
                        <button type="submit" class="btn-primary flex-1 py-3 px-6 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                            Continue to Health Information
                        </button>
                    </div>
                </form>
            </div>
        `;
    }

    showStep4(result) {
        const container = document.getElementById('structured-chat-container');
        container.innerHTML = `
            <div class="step-container" data-step="4">
                <div class="step-header">
                    <h3 class="text-xl font-semibold text-gray-800">${result.question}</h3>
                    <div class="step-progress">
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: 80%"></div>
                        </div>
                        <span class="progress-text">Step 4 of 5</span>
                    </div>
                </div>

                <form class="structured-chat-form mt-6" data-step="4">
                    <div class="form-group">
                        <textarea
                            id="health-situation"
                            name="health_situation"
                            class="structured-input w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            rows="4"
                            placeholder="${result.placeholder}"
                            required></textarea>
                        <div class="field-error text-red-500 text-sm mt-1 hidden"></div>
                    </div>

                    <div class="form-actions mt-6 flex gap-3">
                        <button type="button" class="step-nav-btn btn-secondary px-6 py-3 border border-gray-300 rounded-lg hover:bg-gray-50" data-step="3">
                            Back
                        </button>
                        <button type="submit" class="btn-primary flex-1 py-3 px-6 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                            Continue to Transportation
                        </button>
                    </div>
                </form>
            </div>
        `;
    }

    showStep5(result) {
        const container = document.getElementById('structured-chat-container');
        container.innerHTML = `
            <div class="step-container" data-step="5">
                <div class="step-header">
                    <h3 class="text-xl font-semibold text-gray-800">${result.question}</h3>
                    <div class="step-progress">
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: 100%"></div>
                        </div>
                        <span class="progress-text">Step 5 of 5</span>
                    </div>
                </div>

                <form class="structured-chat-form mt-6" data-step="5">
                    ${result.fields.map((field, index) => `
                        <div class="form-group mb-4">
                            <label for="transport-${index}" class="block text-sm font-medium text-gray-700 mb-2">
                                ${field.label}
                            </label>
                            ${field.options ? `
                                <select
                                    id="transport-${index}"
                                    name="transport_${index}"
                                    class="structured-input w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    ${index === 0 ? 'required' : ''}>
                                    <option value="">Select an option...</option>
                                    ${field.options.map(option => `
                                        <option value="${option}">${option}</option>
                                    `).join('')}
                                </select>
                            ` : `
                                <textarea
                                    id="transport-${index}"
                                    name="transport_${index}"
                                    class="structured-input w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    rows="3"
                                    placeholder="${field.placeholder || ''}"
                                    ${field.label.includes('condition') ? 'required' : ''}></textarea>
                            `}
                        </div>
                    `).join('')}

                    <div class="form-actions mt-6 flex gap-3">
                        <button type="button" class="step-nav-btn btn-secondary px-6 py-3 border border-gray-300 rounded-lg hover:bg-gray-50" data-step="4">
                            Back
                        </button>
                        <button type="submit" class="btn-primary flex-1 py-3 px-6 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors">
                            Get My Recommendation
                        </button>
                    </div>
                </form>
            </div>
        `;
    }

    showRecommendation(result) {
        const container = document.getElementById('structured-chat-container');
        const recommendation = result.recommendation;

        const recommendationColor = this.getRecommendationColor(recommendation.recommendation);

        container.innerHTML = `
            <div class="recommendation-container">
                <div class="recommendation-header text-center mb-8">
                    <div class="recommendation-badge ${recommendationColor} inline-block px-6 py-3 rounded-full text-white font-bold text-lg mb-4">
                        ${recommendation.recommendation}
                    </div>
                    <h3 class="text-2xl font-semibold text-gray-800">Your Purchase Decision Analysis</h3>
                </div>

                <div class="recommendation-content space-y-6">
                    <div class="recommendation-section bg-white border border-gray-200 rounded-lg p-6">
                        <h4 class="text-lg font-semibold text-gray-800 mb-3">💡 Reasoning</h4>
                        <p class="text-gray-700 leading-relaxed">${recommendation.reasoning}</p>
                    </div>

                    ${recommendation.timing ? `
                        <div class="recommendation-section bg-blue-50 border border-blue-200 rounded-lg p-6">
                            <h4 class="text-lg font-semibold text-blue-800 mb-3">⏰ Optimal Timing</h4>
                            <p class="text-blue-700">${recommendation.timing}</p>
                        </div>
                    ` : ''}

                    <div class="recommendation-section bg-green-50 border border-green-200 rounded-lg p-6">
                        <h4 class="text-lg font-semibold text-green-800 mb-3">🎯 Alternatives</h4>
                        <p class="text-green-700">${recommendation.alternatives}</p>
                    </div>

                    <div class="recommendation-section bg-yellow-50 border border-yellow-200 rounded-lg p-6">
                        <h4 class="text-lg font-semibold text-yellow-800 mb-3">📊 Financial Impact</h4>
                        <p class="text-yellow-700">${recommendation.financial_impact}</p>
                    </div>

                    ${result.summary ? `
                        <div class="summary-section bg-gray-50 border border-gray-200 rounded-lg p-6">
                            <h4 class="text-lg font-semibold text-gray-800 mb-3">📋 Decision Summary</h4>
                            <div class="grid md:grid-cols-2 gap-4 text-sm">
                                <div><strong>Item:</strong> ${result.summary.item}</div>
                                <div><strong>Price:</strong> ${result.summary.price}</div>
                                <div><strong>Classification:</strong> ${result.summary.classification}</div>
                                <div><strong>Monthly Income:</strong> $${result.summary.monthly_income}</div>
                            </div>
                        </div>
                    ` : ''}

                    <div class="confidence-indicator mb-6">
                        <div class="flex items-center justify-between mb-2">
                            <span class="text-sm font-medium text-gray-700">Confidence Level</span>
                            <span class="text-sm text-gray-600">${Math.round((recommendation.confidence || 0.85) * 100)}%</span>
                        </div>
                        <div class="confidence-bar bg-gray-200 rounded-full h-2">
                            <div class="confidence-fill bg-blue-500 h-2 rounded-full" style="width: ${(recommendation.confidence || 0.85) * 100}%"></div>
                        </div>
                    </div>
                </div>

                <div class="recommendation-actions mt-8 flex gap-4 justify-center">
                    <button class="restart-flow-btn btn-secondary px-6 py-3 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
                        Analyze Another Purchase
                    </button>
                    <button class="save-decision-btn btn-primary px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                        Save This Decision
                    </button>
                </div>
            </div>
        `;

        // Add save decision handler
        document.querySelector('.save-decision-btn')?.addEventListener('click', () => {
            this.saveDecision(result);
        });
    }

    getRecommendationColor(recommendation) {
        switch (recommendation?.toUpperCase()) {
            case 'BUY':
                return 'bg-green-500';
            case 'WAIT':
                return 'bg-yellow-500';
            case "DON'T BUY":
                return 'bg-red-500';
            default:
                return 'bg-blue-500';
        }
    }

    navigateToStep(targetStep) {
        if (targetStep < 1 || targetStep > 5) return;

        // Simple navigation - in production you'd want to preserve state
        switch (targetStep) {
            case 1:
                this.showStep1();
                break;
            default:
                // For now, restart the flow if trying to go to steps we haven't reached
                this.showStep1();
        }
    }

    restartFlow() {
        this.sessionId = this.generateSessionId();
        this.currentStep = 1;
        this.userData = {};
        this.showStep1();
    }

    async saveDecision(result) {
        try {
            const csrfToken = document.querySelector('meta[name="csrf-token"]')?.content;

            const response = await fetch('/api/decisions/save', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken || ''
                },
                body: JSON.stringify({
                    session_id: this.sessionId,
                    recommendation: result.recommendation,
                    summary: result.summary
                })
            });

            if (response.ok) {
                this.showSuccessMessage('Decision saved successfully!');
            } else {
                this.showError('Failed to save decision. Please try again.');
            }
        } catch (error) {
            console.error('Save decision error:', error);
            this.showError('Failed to save decision. Please try again.');
        }
    }

    validateField(field) {
        const value = field.value.trim();
        const errorElement = field.parentElement.querySelector('.field-error');

        let isValid = true;
        let errorMessage = '';

        // Basic validation
        if (field.hasAttribute('required') && !value) {
            isValid = false;
            errorMessage = 'This field is required.';
        } else if (field.type === 'number' && value && isNaN(value)) {
            isValid = false;
            errorMessage = 'Please enter a valid number.';
        } else if (field.type === 'number' && value && parseFloat(value) < 0) {
            isValid = false;
            errorMessage = 'Please enter a positive number.';
        }

        // Update field styling and error message
        if (isValid) {
            field.classList.remove('border-red-500');
            field.classList.add('border-gray-300');
            errorElement?.classList.add('hidden');
        } else {
            field.classList.remove('border-gray-300');
            field.classList.add('border-red-500');
            if (errorElement) {
                errorElement.textContent = errorMessage;
                errorElement.classList.remove('hidden');
            }
        }

        return isValid;
    }

    showLoadingState(form) {
        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = true;
            submitBtn.innerHTML = `
                <div class="flex items-center justify-center">
                    <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    Processing...
                </div>
            `;
        }
    }

    hideLoadingState(form) {
        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = false;
            // Restore original button text based on step
            const step = parseInt(form.dataset.step);
            const buttonTexts = {
                1: 'Continue to Need vs Want Analysis',
                2: 'Continue to Financial Information',
                3: 'Continue to Health Information',
                4: 'Continue to Transportation',
                5: 'Get My Recommendation'
            };
            submitBtn.innerHTML = buttonTexts[step] || 'Continue';
        }
    }

    showError(message) {
        // Create or update error notification
        let errorDiv = document.getElementById('structured-chat-error');

        if (!errorDiv) {
            errorDiv = document.createElement('div');
            errorDiv.id = 'structured-chat-error';
            errorDiv.className = 'fixed top-4 right-4 bg-red-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
            document.body.appendChild(errorDiv);
        }

        errorDiv.textContent = message;
        errorDiv.classList.remove('hidden');

        // Auto-hide after 5 seconds
        setTimeout(() => {
            errorDiv.classList.add('hidden');
        }, 5000);
    }

    showSuccessMessage(message) {
        // Create or update success notification
        let successDiv = document.getElementById('structured-chat-success');

        if (!successDiv) {
            successDiv = document.createElement('div');
            successDiv.id = 'structured-chat-success';
            successDiv.className = 'fixed top-4 right-4 bg-green-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
            document.body.appendChild(successDiv);
        }

        successDiv.textContent = message;
        successDiv.classList.remove('hidden');

        // Auto-hide after 3 seconds
        setTimeout(() => {
            successDiv.classList.add('hidden');
        }, 3000);
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Only initialize if we're on a page with the structured chat container
    if (document.getElementById('structured-chat-container')) {
        window.structuredChatFlow = new StructuredChatFlow();
    }
});

// Global utility functions for external use
window.StructuredChatUtils = {
    formatCurrency: (amount) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD'
        }).format(amount);
    },

    formatDate: (date) => {
        return new Intl.DateTimeFormat('en-US').format(new Date(date));
    },

    validateFinancialInput: (value) => {
        const num = parseFloat(value);
        return !isNaN(num) && num >= 0;
    }
};