-- This SQL script creates the initial database schema for a task management application.
-- It includes tables for users and tasks, with necessary constraints and indexes.

/* -------------------------------------------------------------------------- */
/*                                   TABLES                                   */
/* -------------------------------------------------------------------------- */
-- Create table for users
-- This table will store user information, including username, email, password hash, and timestamps
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Create table for tasks
-- This table will store tasks for users, including subtasks and kanban board features
CREATE TABLE IF NOT EXISTS tasks (
      id SERIAL PRIMARY KEY,
      user_id INT NOT NULL,
      parent_id INT, -- Recursive relationship for subtasks
      title VARCHAR(255) NOT NULL,
      description TEXT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      due_date TIMESTAMP,
      is_completed BOOLEAN DEFAULT FALSE,
      status VARCHAR(50) DEFAULT 'todo', -- 'todo', 'in-progress', 'done' for kanban
      priority VARCHAR(50) DEFAULT 'low', -- 'low', 'medium', 'high' for kanban
      tags VARCHAR(255)[], -- Array of tags for categorization
      display_order INT DEFAULT 0, -- For ordering tasks in the kanban board
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
      FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

/* -------------------------------------------------------------------------- */
/*                                   INDEXES                                  */
/* -------------------------------------------------------------------------- */
-- Create indexes for faster lookups
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);

/* -------------------------------------------------------------------------- */
/*                                 CONSTRAINS                                 */
/* -------------------------------------------------------------------------- */
-- Add a unique constraint to the username column
ALTER TABLE users ADD CONSTRAINT unique_username UNIQUE (username);
-- Add a unique constraint to the email column
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);

-- Add a unique constraint to the title column in tasks table
ALTER TABLE tasks ADD CONSTRAINT unique_title UNIQUE (title);
-- Add a unique constraint to the user_id and title columns in tasks table
ALTER TABLE tasks ADD CONSTRAINT unique_user_title UNIQUE (user_id, title);
-- Add a unique constraint to the user_id and status columns in tasks table
ALTER TABLE tasks ADD CONSTRAINT unique_user_status UNIQUE (user_id, status);

/* -------------------------------------------------------------------------- */
/*                                  TRIGGERS                                  */
/* -------------------------------------------------------------------------- */
-- Create a trigger to update the updated_at column on each update
-- This trigger will automatically set the updated_at column to the current timestamp whenever a row is updated
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for the users and tasks tables to call the update_updated_at_column function
-- This trigger will automatically set the updated_at column to the current timestamp whenever a row is updated
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create a trigger for the tasks table to call the update_updated_at_column function
-- This trigger will automatically set the updated_at column to the current timestamp whenever a row is updated
CREATE TRIGGER update_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();