create table airports
(
    id               uuid      default gen_random_uuid() not null
        primary key,
    created_at       timestamp default now()             not null,
    modified_at      timestamp default now()             not null,
    name             text                                not null,
    icao             text                                not null
        unique,
    country_code     text                                not null,
    iata             varchar(3),
    state            text,
    country_name     text,
    city             text,
    latitude         double precision,
    longitude        double precision,
    elevation        double precision,
    size             integer,
    is_military      boolean   default false,
    has_lights       boolean   default false,
    is_basecamp      boolean   default false,
    map_surface_type integer,
    is_in_simbrief   boolean   default false,
    display_name     text
);

