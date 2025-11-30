-- Create the 'Restaurants' table
CREATE TABLE Restaurants (
    restaurant_id SERIAL PRIMARY KEY, -- Use SERIAL for auto-incrementing integer
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(50) NOT NULL,
    zip_code VARCHAR(10),
    phone_number VARCHAR(20),
    cuisine_type VARCHAR(50) NOT NULL, -- Replaced ENUM with VARCHAR
    opening_hours VARCHAR(255),  -- e.g., "11:00 AM - 10:00 PM"
    rating DECIMAL(3, 2)         -- e.g., 4.5
);

---

-- Create the 'Customers' table
CREATE TABLE Customers (
    customer_id SERIAL PRIMARY KEY, -- Use SERIAL for auto-incrementing integer
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(50),
    zip_code VARCHAR(10),
    registration_date TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP -- Changed DATETIME to TIMESTAMP WITHOUT TIME ZONE
);

---

-- Create the 'Orders' table
CREATE TABLE Orders (
    order_id SERIAL PRIMARY KEY, -- Use SERIAL for auto-incrementing integer
    customer_id INT NOT NULL,
    restaurant_id INT NOT NULL,
    order_date TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Changed DATETIME
    total_amount DECIMAL(10, 2) NOT NULL,
    payment_method VARCHAR(50) NOT NULL, -- Replaced ENUM with VARCHAR
    order_status VARCHAR(50) DEFAULT 'Pending', -- Replaced ENUM with VARCHAR
    delivery_address VARCHAR(255),
    FOREIGN KEY (customer_id) REFERENCES Customers(customer_id),
    FOREIGN KEY (restaurant_id) REFERENCES Restaurants(restaurant_id)
);

---

-- Create the 'Menu' table
CREATE TABLE Menu (
    menu_id SERIAL PRIMARY KEY, -- Use SERIAL for auto-incrementing integer
    restaurant_id INT NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category VARCHAR(50) NOT NULL, -- Replaced ENUM with VARCHAR
    available BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (restaurant_id) REFERENCES Restaurants(restaurant_id)
);

---

-- Create the 'Order_Items' table (Junction table for Orders and Menu)
CREATE TABLE Order_Items (
    order_item_id SERIAL PRIMARY KEY, -- Use SERIAL for auto-incrementing integer
    order_id INT NOT NULL,
    menu_id INT NOT NULL,
    quantity INT NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    special_instructions TEXT,
    FOREIGN KEY (order_id) REFERENCES Orders(order_id),
    FOREIGN KEY (menu_id) REFERENCES Menu(menu_id)
);

---

-- Create the 'Reviews' table
CREATE TABLE Reviews (
    review_id SERIAL PRIMARY KEY, -- Use SERIAL for auto-incrementing integer
    restaurant_id INT NOT NULL,
    customer_id INT NOT NULL,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5), -- CHECK constraint remains the same
    review_text TEXT,
    review_date TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Changed DATETIME
    FOREIGN KEY (restaurant_id) REFERENCES Restaurants(restaurant_id),
    FOREIGN KEY (customer_id) REFERENCES Customers(customer_id)
);

---

-- Create the 'Delivery_Drivers' table
CREATE TABLE Delivery_Drivers (
    driver_id SERIAL PRIMARY KEY, -- Use SERIAL for auto-incrementing integer
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    vehicle_type VARCHAR(50) NOT NULL, -- Replaced ENUM with VARCHAR
    license_number VARCHAR(20) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    availability_status VARCHAR(50) DEFAULT 'Available', -- Replaced ENUM with VARCHAR
    join_date DATE,
    termination_date DATE NULL
);
