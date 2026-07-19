DELETE FROM permissions p
USING applications a
WHERE p.application_id = a.id AND a.name = 'TALLO-PLANNING' AND p.name = 'squads:manage';
