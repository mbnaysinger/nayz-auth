CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Usamos um bloco anônimo DO $$ para poder guardar os UUIDs gerados em variáveis temporárias
-- e amarrar todas as tabelas corretamente na mesma transação.
DO $$ 
DECLARE
    v_app_id UUID;
    v_user_id UUID;
    v_role_id UUID;
BEGIN
    -- 1. Cria a Aplicação Root (O próprio Painel Administrativo)
    INSERT INTO applications (name, auth_methods) 
    VALUES ('NAYZ-ID', '{"PASSWORD"}')
    RETURNING id INTO v_app_id;

    -- 2. Cria o Usuário Root (Senha padrão criptografada: admin123)
    -- O pgcrypto do Postgres possui a função nativa crypt() para gerar hashes bcrypt internamente!
    INSERT INTO users (email, username, password_hash) 
    VALUES ('teste@teste.com', 'teste', crypt('admin123', gen_salt('bf', 10)))
    RETURNING id INTO v_user_id;

    -- 3. Cria a Role Superior para o Painel
    INSERT INTO roles (application_id, name) 
    VALUES (v_app_id, 'SUPER_ADMIN')
    RETURNING id INTO v_role_id;

    -- 4. Vincula o Usuário Root à Aplicação Admin
    INSERT INTO user_applications (user_id, application_id) 
    VALUES (v_user_id, v_app_id);

    -- 5. Concede a Role de Super Admin ao Usuário
    INSERT INTO user_roles (user_id, role_id) 
    VALUES (v_user_id, v_role_id);
END $$;
