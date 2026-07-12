-- Removemos as entradas baseadas nos dados exatos do Seed.
-- Graças à restrição ON DELETE CASCADE definida na migração 000001,
-- ao deletarmos a aplicação e o usuário, o banco apagará automaticamente
-- as roles, user_applications e user_roles atreladas a eles!

DELETE FROM users WHERE email = 'root@nayz.tech';
DELETE FROM applications WHERE name = 'Nayz Auth Console';
