-- Cassandra Schema for Game Backend

CREATE KEYSPACE IF NOT EXISTS game_keyspace 
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};

USE game_keyspace;

CREATE TABLE IF NOT EXISTS item_templates (
    template_id text PRIMARY KEY,
    name text,
    type int,
    base_attack double,
    base_defense double,
    rarity text,
    base_price double
);

CREATE TABLE IF NOT EXISTS npcs (
    id text PRIMARY KEY,
    name text,
    map_id text,
    dialogue text,
    shop_item_template_ids list<text>
);

CREATE INDEX IF NOT EXISTS idx_npcs_map_id ON npcs(map_id);

CREATE TABLE IF NOT EXISTS dungeon_layouts (
    map_id text PRIMARY KEY,
    name text,
    width double,
    height double,
    is_safe_zone boolean,
    spawn_x double,
    spawn_y double
);

CREATE TABLE IF NOT EXISTS analytics_events (
    id timeuuid,
    topic text,
    event_key text,
    payload text,
    timestamp timestamp,
    created_at timestamp,
    PRIMARY KEY (topic, id)
) WITH CLUSTERING ORDER BY (id DESC);
