-- Migration: Create expenses table
-- Description: Create table to store user expenses

CREATE TABLE IF NOT EXISTS `expenses` (
    `id` VARCHAR(36) PRIMARY KEY,
    `user_id` VARCHAR(36) NOT NULL,
    `category` VARCHAR(50) NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `amount` DECIMAL(10,2) NOT NULL,
    `frequency` VARCHAR(20) NOT NULL,
    `is_fixed` BOOLEAN NOT NULL DEFAULT FALSE,
    `priority` TINYINT NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    
    INDEX `idx_expenses_user_id` (`user_id`),
    INDEX `idx_expenses_category` (`category`),
    INDEX `idx_expenses_priority` (`priority`),
    INDEX `idx_expenses_is_fixed` (`is_fixed`),
    INDEX `idx_expenses_deleted_at` (`deleted_at`),
    
    CONSTRAINT `fk_expenses_user_id` 
        FOREIGN KEY (`user_id`) 
        REFERENCES `users` (`id`) 
        ON DELETE CASCADE ON UPDATE CASCADE,
        
    CONSTRAINT `chk_expenses_category` 
        CHECK (`category` IN ('housing', 'food', 'transport', 'entertainment', 'utilities', 'other')),
        
    CONSTRAINT `chk_expenses_frequency` 
        CHECK (`frequency` IN ('monthly', 'weekly', 'daily')),
        
    CONSTRAINT `chk_expenses_priority` 
        CHECK (`priority` BETWEEN 1 AND 3),
        
    CONSTRAINT `chk_expenses_amount` 
        CHECK (`amount` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;