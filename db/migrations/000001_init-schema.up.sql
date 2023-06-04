CREATE TABLE "users" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "name" VARCHAR,
    "email" VARCHAR,
    "photo" VARCHAR,
    "cover" VARCHAR,
    "verified" BOOLEAN DEFAULT 'yes',
    "role" VARCHAR NOT NULL DEFAULT 'user',
    "address" VARCHAR NOT NULL,
    "nonce" VARCHAR NOT NULL,
    "status" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "tokens" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "token_index" NUMERIC NOT NULL, 
    "title" VARCHAR NOT NULL,
    "description" VARCHAR,
    "category_id" UUID,
    "collection_id" UUID,
    "image" VARCHAR NOT NULL,
    "uri" VARCHAR NOT NULL,
    "source_id" UUID,
    "fraction_id" UUID,
    "supply" NUMERIC NOT NULL DEFAULT 1,
    "last_price" DOUBLE PRECISION NOT NULL,
    "views" NUMERIC NOT NULL,
    "number_of_transactions" NUMERIC NOT NULL,
    "volume_transactions" NUMERIC NOT NULL,
    "creator_id" UUID NOT NULL,
    "attributes" JSONB,
    "status" VARCHAR NOT NULL,
    "transaction_hash" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "tokens_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "fractions" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "token_parent_id" UUID NOT NULL,
    "token_fraction_id" UUID NOT NULL,
    "status" VARCHAR NOT NULL,
    "transaction_hash" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "fractions_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "token_categories" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "title" VARCHAR NOT NULL,
    "description" VARCHAR NOT NULL,
    "icon" VARCHAR NOT NULL,
    "status" VARCHAR NOT NULL,
    "transaction_hash" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "token_categories_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "collections" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "thumbnail" VARCHAR NOT NULL,
    "cover" VARCHAR NOT NULL,
    "title" VARCHAR NOT NULL,
    "description" VARCHAR NOT NULL,
    "number_of_items" NUMERIC NOT NULL DEFAULT 0,
    "views" NUMERIC NOT NULL,
    "number_of_transactions" NUMERIC NOT NULL,
    "volume_transactions" NUMERIC NOT NULL,
    "floor" NUMERIC NOT NULL DEFAULT 0,
    "category_id" UUID NOT NULL,
    "creator_id" UUID NOT NULL,
    "status" VARCHAR NOT NULL,
    "transaction_hash" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "collections_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "ownerships" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "user_id" UUID,
    "token_id" UUID NOT NULL,
    "quantity" NUMERIC NOT NULL,
    "sale_price" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "initial_price" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "rent_cost" DOUBLE PRECISION DEFAULT 0,
    "available_for_sale" BOOLEAN DEFAULT 'false',
    "available_for_rent" BOOLEAN DEFAULT 'false',
    "status" VARCHAR NOT NULL,
    "transaction_hash" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "ownerships_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "rentals" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "user_id" UUID NOT NULL,
    "owner_id" UUID NOT NULL,
    "token_id" UUID NOT NULL,
    "ownership_id" VARCHAR NOT NULL,
    "timestamp" NUMERIC NOT NULL,
    "status" VARCHAR NOT NULL,
    "transaction_hash" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "rentals_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "transactions" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "user_from_id" UUID NOT NULL,
    "user_to_id" UUID NOT NULL,
    "ownership_id" UUID NOT NULL,
    "rental_id" UUID,
    "token_id" UUID NOT NULL,
    "collection_id" UUID NOT NULL,
    "quantity" NUMERIC NOT NULL,
    "amount" DOUBLE PRECISION NOT NULL,
    "gas_fee" DOUBLE PRECISION NOT NULL,
    "type" VARCHAR NOT NULL DEFAULT 'purchase',
    "status" VARCHAR NOT NULL,
    "transaction_hash" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "transactions_pkey" PRIMARY KEY ("id")
);  