CREATE TABLE IF NOT EXISTS player_inventory (
    id VARCHAR(36) PRIMARY KEY,
    player_id VARCHAR(36) NOT NULL,
    template_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    item_type INT NOT NULL, -- 0: WEAPON, 1: ARMOUR, 2: ARMAMENT
    attack NUMERIC(10, 2) DEFAULT 0,
    defense NUMERIC(10, 2) DEFAULT 0,
    rarity VARCHAR(20) DEFAULT 'Common',
    is_equipped BOOLEAN DEFAULT FALSE,
    price NUMERIC(15, 2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_inventory_player ON player_inventory(player_id);
