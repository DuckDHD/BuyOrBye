-- Migration: Create finance_summaries table
-- Description: Create table to store aggregated financial health summaries for users

CREATE TABLE IF NOT EXISTS `finance_summaries` (
    `user_id` VARCHAR(36) PRIMARY KEY,
    `monthly_income` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `monthly_expenses` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `monthly_loan_payments` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `disposable_income` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `debt_to_income_ratio` DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    `savings_rate` DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    `financial_health` VARCHAR(20) NOT NULL DEFAULT 'Poor',
    `budget_remaining` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    
    INDEX `idx_finance_summaries_financial_health` (`financial_health`),
    INDEX `idx_finance_summaries_updated_at` (`updated_at`),
    INDEX `idx_finance_summaries_deleted_at` (`deleted_at`),
    
    CONSTRAINT `fk_finance_summaries_user_id` 
        FOREIGN KEY (`user_id`) 
        REFERENCES `users` (`id`) 
        ON DELETE CASCADE ON UPDATE CASCADE,
        
    CONSTRAINT `chk_finance_summaries_health` 
        CHECK (`financial_health` IN ('Excellent', 'Good', 'Fair', 'Poor')),
        
    CONSTRAINT `chk_finance_summaries_income` 
        CHECK (`monthly_income` >= 0),
        
    CONSTRAINT `chk_finance_summaries_expenses` 
        CHECK (`monthly_expenses` >= 0),
        
    CONSTRAINT `chk_finance_summaries_loan_payments` 
        CHECK (`monthly_loan_payments` >= 0),
        
    CONSTRAINT `chk_finance_summaries_debt_ratio` 
        CHECK (`debt_to_income_ratio` >= 0),
        
    CONSTRAINT `chk_finance_summaries_savings_rate` 
        CHECK (`savings_rate` >= -1.0000)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;