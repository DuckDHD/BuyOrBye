-- Migration: Create incomes table
-- Description: Create table to store user income sources

CREATE TABLE IF NOT EXISTS `incomes` (
    `id` VARCHAR(36) PRIMARY KEY,
    `user_id` VARCHAR(36) NOT NULL,
    `source` VARCHAR(255) NOT NULL,
    `amount` DECIMAL(10,2) NOT NULL,
    `frequency` VARCHAR(20) NOT NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    
    INDEX `idx_incomes_user_id` (`user_id`),
    INDEX `idx_incomes_deleted_at` (`deleted_at`),
    INDEX `idx_incomes_is_active` (`is_active`),
    
    CONSTRAINT `fk_incomes_user_id` 
        FOREIGN KEY (`user_id`) 
        REFERENCES `users` (`id`) 
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;