CREATE OR REPLACE LANGUAGE plpgsql;

-- Trigger for updating `updated_at` columns:
CREATE OR REPLACE FUNCTION on_update_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = transaction_timestamp();
    RETURN new;
END;
$$ language 'plpgsql';

-- Domain for email addresses:
CREATE EXTENSION IF NOT EXISTS plperl;
CREATE OR REPLACE LANGUAGE plperlu;

CREATE OR REPLACE FUNCTION valid_email(text) RETURNS boolean
LANGUAGE 'plperlu'
IMMUTABLE LEAKPROOF STRICT AS $$
    use Email::Valid;
    my $email = shift;
    Email::Valid->address($email) or die "Invalid email address: $email\n";
    return 'true';
$$;

CREATE DOMAIN validemail AS text NOT NULL
    CONSTRAINT validemail_check CHECK (valid_email(VALUE));

-- Domain for time zones:
CREATE OR REPLACE FUNCTION valid_time_zone(tz text) RETURNS BOOLEAN
LANGUAGE 'plpgsql'
IMMUTABLE STRICT AS $$
DECLARE date timestamptz;
BEGIN
    date := now() AT TIME ZONE tz;
    RETURN true;
EXCEPTION WHEN OTHERS THEN
    RETURN FALSE;
END;
$$;

CREATE DOMAIN timezone AS text NOT NULL
    CONSTRAINT timezone_check CHECK (valid_time_zone(VALUE));
