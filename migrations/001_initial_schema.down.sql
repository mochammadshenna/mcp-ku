-- Drop triggers
DROP TRIGGER IF EXISTS update_tools_updated_at ON tools;
DROP TRIGGER IF EXISTS update_evaluations_updated_at ON evaluations;
DROP TRIGGER IF EXISTS update_prompts_updated_at ON prompts;
DROP TRIGGER IF EXISTS update_flows_updated_at ON flows;
DROP TRIGGER IF EXISTS update_mcp_servers_updated_at ON mcp_servers;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_tool_executions_request_id;
DROP INDEX IF EXISTS idx_tool_executions_success;
DROP INDEX IF EXISTS idx_tool_executions_tool_id;
DROP INDEX IF EXISTS idx_tools_handler;
DROP INDEX IF EXISTS idx_tools_name;
DROP INDEX IF EXISTS idx_evaluation_results_score;
DROP INDEX IF EXISTS idx_evaluation_results_generation_id;
DROP INDEX IF EXISTS idx_evaluation_results_evaluation_id;
DROP INDEX IF EXISTS idx_evaluations_name;
DROP INDEX IF EXISTS idx_evaluations_type;
DROP INDEX IF EXISTS idx_generations_created_at;
DROP INDEX IF EXISTS idx_generations_request_id;
DROP INDEX IF EXISTS idx_generations_status;
DROP INDEX IF EXISTS idx_generations_model;
DROP INDEX IF EXISTS idx_vector_documents_created_at;
DROP INDEX IF EXISTS idx_vector_documents_source;
DROP INDEX IF EXISTS idx_vector_documents_embedding;
DROP INDEX IF EXISTS idx_prompts_version;
DROP INDEX IF EXISTS idx_prompts_name;
DROP INDEX IF EXISTS idx_flow_executions_started_at;
DROP INDEX IF EXISTS idx_flow_executions_status;
DROP INDEX IF EXISTS idx_flow_executions_flow_id;
DROP INDEX IF EXISTS idx_flows_created_at;
DROP INDEX IF EXISTS idx_flows_name;
DROP INDEX IF EXISTS idx_mcp_servers_created_at;
DROP INDEX IF EXISTS idx_mcp_servers_status;

-- Drop tables
DROP TABLE IF EXISTS tool_executions;
DROP TABLE IF EXISTS tools;
DROP TABLE IF EXISTS evaluation_results;
DROP TABLE IF EXISTS evaluations;
DROP TABLE IF EXISTS generations;
DROP TABLE IF EXISTS vector_documents;
DROP TABLE IF EXISTS prompts;
DROP TABLE IF EXISTS flow_executions;
DROP TABLE IF EXISTS flows;
DROP TABLE IF EXISTS mcp_servers;

-- Drop extension (be careful with this in production)
-- DROP EXTENSION IF EXISTS vector;