create table if not exists otpbcc
(
    id serial,
    fail_count bigint,
    otp_status text,
    activity_id bigint,
    phone text not null ,
    created_at bigint,
    updated_at bigint,
    unique  (phone,activity_id)
);