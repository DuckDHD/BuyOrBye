-- Migration: Create loans table
-- Description: Create table to store user loans

CREATE TABLE IF NOT EXISTS `loans` (
    `id` VARCHAR(36) PRIMARY KEY,
    `user_id` VARCHAR(36) NOT NULL,
    `lender` VARCHAR(255) NOT NULL,
    `type` VARCHAR(50) NOT NULL,
    `principal_amount` DECIMAL(12,2) NOT NULL,
    `remaining_balance` DECIMAL(12,2) NOT NULL,
    `monthly_payment` DECIMAL(10,2) NOT NULL,
    `interest_rate` DECIMAL(5,3) NOT NULL,
    `end_date` DATE NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    
    INDEX `idx_loans_user_id` (`user_id`),
    INDEX `idx_loans_type` (`type`),
    INDEX `idx_loans_end_date` (`end_date`),
    INDEX `idx_loans_deleted_at` (`deleted_at`),
    
    CONSTRAINT `fk_loans_user_id` 
        FOREIGN KEY (`user_id`) 
        REFERENCES `users` (`id`) 
        ON DELETE CASCADE ON UPDATE CASCADE,
        
    CONSTRAINT `chk_loans_type` 
        CHECK (`type` IN ('mortgage', 'auto', 'personal', 'student')),
        
    CONSTRAINT `chk_loans_principal_amount` 
        CHECK (`principal_amount` > 0),
        
    CONSTRAINT `chk_loans_remaining_balance` 
        CHECK (`remaining_balance` >= 0),
        
    CONSTRAINT `chk_loans_monthly_payment` 
        CHECK (`monthly_payment` > 0),
        
    CONSTRAINT `chk_loans_interest_rate` 
        CHECK (`interest_rate` >= 0 AND `interest_rate` <= 100),
        
    CONSTRAINT `chk_loans_balance_vs_principal` 
        CHECK (`remaining_balance` <= `principal_amount`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;