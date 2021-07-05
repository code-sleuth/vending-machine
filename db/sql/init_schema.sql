CREATE TABLE IF NOT EXISTS "users" (
    "uuid" VARCHAR(50) PRIMARY KEY,
    "username" VARCHAR(50) UNIQUE NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    "deposit" INTEGER NOT NULL,
    "role" VARCHAR(10) NOT NULL
);

CREATE TABLE IF NOT EXISTS "products" (
    "uuid" VARCHAR(50) PRIMARY KEY,
    "amount_available" INTEGER NOT NULL,
    "cost" INTEGER NOT NULL,
    "product_name" VARCHAR(255) NOT NULL,
    "seller_id" VARCHAR(50) NOT NULL
);

DO $$
BEGIN
    BEGIN
        ALTER TABLE "products" ADD FOREIGN KEY ("seller_id") REFERENCES "users" ("uuid") ON DELETE CASCADE;
    EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'Table foreign key products.seller_id already exists';
    END;
END $$;

DO $$
BEGIN
    BEGIN
        ALTER TABLE "users" ADD CONSTRAINT "username" UNIQUE ("username");
    EXCEPTION
        WHEN duplicate_table THEN RAISE NOTICE 'Table constraint users.username already exists';
    END;

END $$;

