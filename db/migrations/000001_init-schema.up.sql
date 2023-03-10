CREATE TABLE "users" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "name" VARCHAR,
    "email" VARCHAR,
    "photo" VARCHAR,
    "verified" BOOLEAN DEFAULT 'yes',
    "role" VARCHAR NOT NULL DEFAULT 'user',
    "address" VARCHAR NOT NULL,
    "nonce" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "tokens" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "title" VARCHAR NOT NULL,
    "description" VARCHAR,
    "category_id" VARCHAR NOT NULL,
    "collection_id" VARCHAR,
    "image" VARCHAR NOT NULL,
    "uri" VARCHAR NOT NULL,
    "fraction_id" VARCHAR,
    "quantity" NUMERIC NOT NULL DEFAULT 1,
    "last_price" NUMERIC NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "tokens_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "fractions" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "token_parent_id" VARCHAR NOT NULL,
    "token_fraction_id" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "fractions_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "token_categories" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "title" VARCHAR NOT NULL,
    "description" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "token_categories_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "collections" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "title" VARCHAR NOT NULL,
    "description" VARCHAR NOT NULL,
    "user_id" VARCHAR NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "collections_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "ownerships" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "user_id" VARCHAR NOT NULL,
    "token_id" VARCHAR NOT NULL,
    "quantity" VARCHAR NOT NULL,
    "sale_price" NUMERIC NOT NULL DEFAULT 0,
    "rent_cost" NUMERIC DEFAULT 0,
    "available_for_sale" BOOLEAN DEFAULT 'false',
    "available_for_rent" BOOLEAN DEFAULT 'false',
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "ownerships_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "rentals" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "user_id" VARCHAR NOT NULL,
    "token_id" VARCHAR NOT NULL,
    "timestamp" TIMESTAMP(3) NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "rentals_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "transactions" (
    "id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
    "user_from_id" VARCHAR NOT NULL,
    "user_to_id" VARCHAR NOT NULL,
    "token_id" VARCHAR NOT NULL,
    "quantity" NUMERIC NOT NULL,
    "price" NUMERIC NOT NULL,
    "gass_fee" NUMERIC NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "transactions_pkey" PRIMARY KEY ("id")
);