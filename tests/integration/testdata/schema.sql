CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	name text NOT NULL
);

CREATE TABLE posts (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title text NOT NULL,
	published boolean NOT NULL DEFAULT false
);

CREATE TABLE tags (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	name text NOT NULL UNIQUE
);

CREATE TABLE post_tags (
	post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
	tag_id uuid NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
	PRIMARY KEY (post_id, tag_id)
);
