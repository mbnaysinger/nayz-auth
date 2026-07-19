DELETE FROM roles r
USING applications a
WHERE r.application_id = a.id AND a.name = 'TALLO-PLANNING' AND r.name = 'USER-FULL';
