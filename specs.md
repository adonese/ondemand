# Service on demand app


- user
- service provider

- user apis

    - registration
    - login
    - authentication
    - password retrieval
    - user profile
    - list of services
    - list of service features
    - orders
    - order inquiries
    - history
    - completion and rating

- service provider apis

    - registration
    - login
    - authentication
    - password retrieval
    - user profile
    - provided services
    - orders
    - history

- administrator
    - services CRUD
    - service providers
    - service providers screening
    - payments
    - outstanding orders
    - issues


### Service provider sign up logic

- registration_service
- is_active: false
- admin_acceptance
- email_prompt


### user sign up

- user_registration_service
- email_confirmation

# Views

- sign_up
- login
- authentication
- profile
- list_services
- request_service


# orders

- list_orders
- order_status
- made_by
- to_whom
- date


# database models

## User profile

- username
- is_admin
- is_service_provider
- firstname
- lastname
- birthday
- location (gps) optional
- address
- on_job (for service provider)

## services (lookup table)

- service_name
- service_id

## orders

- user
- service_id
- service_provider_id
- is_completed
- user_rating
- service_provider_rating
- order_id

## order criteria

- geolocation fetch for nearest available user
