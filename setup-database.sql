-- Order Processing System Database Setup
-- Database: order_system
-- User: orderuser

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255),
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(100),
    transaction_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create payment_transactions table (required by payment-processing-service)
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    transaction_id VARCHAR(255),
    message TEXT,
    provider_response TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create users table (required by auth-service)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'customer',
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create user_sessions table (required by auth-service)
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    refresh_token_hash VARCHAR(255) UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_agent TEXT,
    ip_address INET
);

-- Create order_status_history table (required by order-status-service)
CREATE TABLE IF NOT EXISTS order_status_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create shipments table
CREATE TABLE IF NOT EXISTS shipments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    tracking_number VARCHAR(255) UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'preparing',
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create stock_reservations table
CREATE TABLE IF NOT EXISTS stock_reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID REFERENCES products(id),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'reserved',
    reserved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample products with specific UUIDs - SIMPLIFIED VERSION
-- First batch: Electronics
INSERT INTO products (id, name, description, price, stock_quantity) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'iPhone 15 Pro Max 256GB', 'Apple iPhone 15 Pro Max, 256GB depolama, Pro kamera sistemi', 54999.99, 25),
('550e8400-e29b-41d4-a716-446655440002', 'Samsung Galaxy S24 Ultra', 'Samsung Galaxy S24 Ultra, 512GB, S Pen dahil', 52999.99, 30),
('550e8400-e29b-41d4-a716-446655440003', 'MacBook Pro 16" M3', 'Apple MacBook Pro 16", M3 chip, 32GB RAM, 1TB SSD', 89999.99, 15),
('550e8400-e29b-41d4-a716-446655440004', 'Dell XPS 13 Plus', 'Dell XPS 13 Plus, Intel i7, 16GB RAM, 512GB SSD', 45999.99, 20),
('550e8400-e29b-41d4-a716-446655440005', 'iPad Pro 12.9" M4', 'Apple iPad Pro 12.9", M4 chip, 256GB', 39999.99, 18),
('550e8400-e29b-41d4-a716-446655440006', 'Surface Pro 9', 'Microsoft Surface Pro 9, Intel i7, 16GB RAM', 42999.99, 22),
('550e8400-e29b-41d4-a716-446655440007', 'AirPods Pro 2. Nesil', 'Apple AirPods Pro 2. nesil, USB-C şarj kutusu', 8999.99, 50),
('550e8400-e29b-41d4-a716-446655440008', 'Sony WH-1000XM5', 'Sony WH-1000XM5 kablosuz kulaklık, aktif gürültü engelleme', 12999.99, 35),
('550e8400-e29b-41d4-a716-446655440009', 'Apple Watch Series 9 45mm', 'Apple Watch Series 9, 45mm, GPS + Cellular', 17999.99, 28),
('550e8400-e29b-41d4-a716-446655440010', 'Samsung Galaxy Watch 6', 'Samsung Galaxy Watch 6, 44mm, Bluetooth', 9999.99, 40),
('550e8400-e29b-41d4-a716-446655440011', 'Canon EOS R5', 'Canon EOS R5 fotoğraf makinesi, 45MP full frame', 159999.99, 8),
('550e8400-e29b-41d4-a716-446655440012', 'Sony Alpha A7 IV', 'Sony Alpha A7 IV aynasız fotoğraf makinesi, 33MP', 129999.99, 12),
('550e8400-e29b-41d4-a716-446655440013', 'GoPro Hero 12 Black', 'GoPro Hero 12 Black aksiyon kamerası, 5.3K video', 19999.99, 45),
('550e8400-e29b-41d4-a716-446655440014', 'DJI Mini 4 Pro', 'DJI Mini 4 Pro drone, 4K HDR video, 34 dakika uçuş', 39999.99, 15),
('550e8400-e29b-41d4-a716-446655440015', 'Nintendo Switch OLED', 'Nintendo Switch OLED model, 64GB dahili hafıza', 12999.99, 60),
('550e8400-e29b-41d4-a716-446655440016', 'PlayStation 5', 'Sony PlayStation 5 oyun konsolu, 825GB SSD', 24999.99, 10),
('550e8400-e29b-41d4-a716-446655440017', 'Xbox Series X', 'Microsoft Xbox Series X, 1TB SSD, 4K gaming', 22999.99, 12),
('550e8400-e29b-41d4-a716-446655440018', 'Meta Quest 3 128GB', 'Meta Quest 3 VR başlığı, 128GB depolama', 19999.99, 20),
('550e8400-e29b-41d4-a716-446655440019', 'Tesla Model S Plaid Wheel', 'Tesla Model S Plaid direksiyon simülatörü', 4999.99, 8),
('550e8400-e29b-41d4-a716-446655440020', 'Logitech MX Master 3S', 'Logitech MX Master 3S kablosuz fare, ergonomik', 3999.99, 80),
('550e8400-e29b-41d4-a716-446655440021', 'Corsair K95 RGB Platinum', 'Corsair K95 RGB Platinum mekanik klavye', 8999.99, 35),
('550e8400-e29b-41d4-a716-446655440022', 'LG UltraWide 34"', 'LG UltraWide 34" monitör, 3440x1440, 144Hz', 19999.99, 18),
('550e8400-e29b-41d4-a716-446655440023', 'Samsung Odyssey G9 49"', 'Samsung Odyssey G9 49" curved gaming monitör', 39999.99, 5),
('550e8400-e29b-41d4-a716-446655440024', 'ASUS ROG Strix RTX 4090', 'ASUS ROG Strix GeForce RTX 4090, 24GB GDDR6X', 79999.99, 6),
('550e8400-e29b-41d4-a716-446655440025', 'AMD Ryzen 9 7950X', 'AMD Ryzen 9 7950X işlemci, 16 çekirdek', 24999.99, 15),

-- Giyim ve Moda (101-200)
('550e8400-e29b-41d4-a716-446655440101', 'Nike Air Max 270', 'Nike Air Max 270 spor ayakkabı, siyah/beyaz', 4999.99, 120),
('550e8400-e29b-41d4-a716-446655440102', 'Adidas Ultraboost 23', 'Adidas Ultraboost 23 koşu ayakkabısı', 6999.99, 85),
('550e8400-e29b-41d4-a716-446655440103', 'Levis 501 Original Jean', 'Levis 501 Original kot pantolon, klasik kesim', 2999.99, 200),
('550e8400-e29b-41d4-a716-446655440104', 'Zara Oversize Blazer', 'Zara oversize blazer ceket, lacivert', 1999.99, 50),
('550e8400-e29b-41d4-a716-446655440105', 'H&M Basic T-Shirt', 'H&M basic pamuklu t-shirt, çeşitli renkler', 299.99, 500),
('550e8400-e29b-41d4-a716-446655440106', 'Mango Midi Elbise', 'Mango midi boy elbise, çiçek desenli', 1499.99, 75),
('550e8400-e29b-41d4-a716-446655440107', 'Tommy Hilfiger Polo', 'Tommy Hilfiger polo t-shirt, %100 pamuk', 1799.99, 90),
('550e8400-e29b-41d4-a716-446655440108', 'Calvin Klein Boxer 3lü', 'Calvin Klein boxer 3lü set, pamuklu', 899.99, 150),
('550e8400-e29b-41d4-a716-446655440109', 'Victoria Secret Push-up', 'Victoria Secret push-up sütyen', 1299.99, 60),
('550e8400-e29b-41d4-a716-446655440110', 'Ray-Ban Aviator', 'Ray-Ban Aviator güneş gözlüğü, altın çerçeve', 5999.99, 40),
('550e8400-e29b-41d4-a716-446655440111', 'Rolex Submariner', 'Rolex Submariner su altı saati, otomatik', 399999.99, 2),
('550e8400-e29b-41d4-a716-446655440112', 'Casio G-Shock', 'Casio G-Shock dijital saat, şok dayanıklı', 1999.99, 80),
('550e8400-e29b-41d4-a716-446655440113', 'Michael Kors Çanta', 'Michael Kors deri el çantası, siyah', 4999.99, 25),
('550e8400-e29b-41d4-a716-446655440114', 'Louis Vuitton Neverfull', 'Louis Vuitton Neverfull tote çanta', 59999.99, 3),
('550e8400-e29b-41d4-a716-446655440115', 'Nike Air Force 1', 'Nike Air Force 1 beyaz spor ayakkabı', 3999.99, 150),
('550e8400-e29b-41d4-a716-446655440116', 'Converse Chuck Taylor', 'Converse Chuck Taylor All Star, siyah', 2499.99, 180),
('550e8400-e29b-41d4-a716-446655440117', 'Puma RS-X', 'Puma RS-X retro spor ayakkabı', 3499.99, 70),
('550e8400-e29b-41d4-a716-446655440118', 'New Balance 990v5', 'New Balance 990v5 koşu ayakkabısı', 7999.99, 45),
('550e8400-e29b-41d4-a716-446655440119', 'Vans Old Skool', 'Vans Old Skool skate ayakkabısı', 2799.99, 120),
('550e8400-e29b-41d4-a716-446655440120', 'Under Armour Hoodie', 'Under Armour kapüşonlu sweatshirt', 2499.99, 95),

-- Ev ve Yaşam (201-300)
('550e8400-e29b-41d4-a716-446655440201', 'IKEA MALM Yatak Odası Seti', 'IKEA MALM yatak odası takımı, beyaz', 12999.99, 20),
('550e8400-e29b-41d4-a716-446655440202', 'Samsung 85" QLED TV', 'Samsung 85" 4K QLED Smart TV, HDR10+', 89999.99, 8),
('550e8400-e29b-41d4-a716-446655440203', 'LG OLED 77" C3', 'LG OLED 77" C3 4K Smart TV', 79999.99, 10),
('550e8400-e29b-41d4-a716-446655440204', 'Dyson V15 Detect', 'Dyson V15 Detect kablosuz süpürge', 19999.99, 35),
('550e8400-e29b-41d4-a716-446655440205', 'iRobot Roomba j7+', 'iRobot Roomba j7+ robot süpürge', 24999.99, 25),
('550e8400-e29b-41d4-a716-446655440206', 'Nespresso Vertuo Next', 'Nespresso Vertuo Next kapsül kahve makinesi', 4999.99, 50),
('550e8400-e29b-41d4-a716-446655440207', 'Breville Barista Express', 'Breville Barista Express espresso makinesi', 19999.99, 15),
('550e8400-e29b-41d4-a716-446655440208', 'KitchenAid Stand Mixer', 'KitchenAid Stand Mixer hamur karıştırıcı', 14999.99, 20),
('550e8400-e29b-41d4-a716-446655440209', 'Instant Pot Duo 7-in-1', 'Instant Pot Duo 7-in-1 çok fonksiyonlu tencere', 3999.99, 60),
('550e8400-e29b-41d4-a716-446655440210', 'Ninja Foodi Air Fryer', 'Ninja Foodi hava fritözü, 8 quart', 7999.99, 40),
('550e8400-e29b-41d4-a716-446655440211', 'Weber Genesis II E-335', 'Weber Genesis II E-335 gaz barbekü', 29999.99, 8),
('550e8400-e29b-41d4-a716-446655440212', 'Tempur-Pedic Yatak', 'Tempur-Pedic memory foam yatak, çift kişilik', 39999.99, 12),
('550e8400-e29b-41d4-a716-446655440213', 'Herman Miller Aeron', 'Herman Miller Aeron ergonomik ofis koltuğu', 49999.99, 6),
('550e8400-e29b-41d4-a716-446655440214', 'IKEA BEKANT Çalışma Masası', 'IKEA BEKANT ayarlanabilir çalışma masası', 4999.99, 30),
('550e8400-e29b-41d4-a716-446655440215', 'Philips Hue Akıllı Ampul', 'Philips Hue akıllı LED ampul, RGB', 1499.99, 100),
('550e8400-e29b-41d4-a716-446655440216', 'Ring Video Doorbell 4', 'Ring Video Doorbell 4 akıllı kapı zili', 7999.99, 35),
('550e8400-e29b-41d4-a716-446655440217', 'Nest Learning Thermostat', 'Google Nest Learning akıllı termostat', 8999.99, 25),
('550e8400-e29b-41d4-a716-446655440218', 'Amazon Echo Dot 5. Nesil', 'Amazon Echo Dot 5. nesil akıllı hoparlör', 1999.99, 80),
('550e8400-e29b-41d4-a716-446655440219', 'Sonos Arc Soundbar', 'Sonos Arc premium soundbar, Dolby Atmos', 32999.99, 15),
('550e8400-e29b-41d4-a716-446655440220', 'Bose SoundLink Revolve+', 'Bose SoundLink Revolve+ taşınabilir hoparlör', 9999.99, 45),

-- Spor ve Outdoor (301-400)
('550e8400-e29b-41d4-a716-446655440301', 'Peloton Bike+', 'Peloton Bike+ akıllı exercise bisikleti', 89999.99, 5),
('550e8400-e29b-41d4-a716-446655440302', 'NordicTrack Treadmill', 'NordicTrack Commercial 1750 koşu bandı', 59999.99, 8),
('550e8400-e29b-41d4-a716-446655440303', 'Bowflex SelectTech 552', 'Bowflex SelectTech 552 ayarlanabilir dumbbell', 19999.99, 20),
('550e8400-e29b-41d4-a716-446655440304', 'Garmin Forerunner 955', 'Garmin Forerunner 955 GPS koşu saati', 19999.99, 25),
('550e8400-e29b-41d4-a716-446655440305', 'Polar Vantage V2', 'Polar Vantage V2 multisport watch', 17999.99, 30),
('550e8400-e29b-41d4-a716-446655440306', 'YETI Rambler 30 oz', 'YETI Rambler 30 oz paslanmaz çelik tumbler', 1499.99, 150),
('550e8400-e29b-41d4-a716-446655440307', 'Hydro Flask 32 oz', 'Hydro Flask 32 oz wide mouth su şişesi', 1299.99, 200),
('550e8400-e29b-41d4-a716-446655440308', 'Patagonia Better Sweater', 'Patagonia Better Sweater polar ceket', 3999.99, 60),
('550e8400-e29b-41d4-a716-446655440309', 'The North Face Resolve 2', 'The North Face Resolve 2 yağmurluk', 2999.99, 80),
('550e8400-e29b-41d4-a716-446655440310', 'Columbia Flash Forward', 'Columbia Flash Forward outdoor pantolon', 2499.99, 100),
('550e8400-e29b-41d4-a716-446655440311', 'REI Co-op Merino Wool', 'REI Co-op merino yün base layer', 1999.99, 120),
('550e8400-e29b-41d4-a716-446655440312', 'Osprey Atmos AG 65', 'Osprey Atmos AG 65 trekking sırt çantası', 12999.99, 15),
('550e8400-e29b-41d4-a716-446655440313', 'Deuter Speed Lite 20', 'Deuter Speed Lite 20 hiking sırt çantası', 2999.99, 50),
('550e8400-e29b-41d4-a716-446655440314', 'Black Diamond Spot 400', 'Black Diamond Spot 400 kafa lambası', 1499.99, 80),
('550e8400-e29b-41d4-a716-446655440315', 'MSR PocketRocket 2', 'MSR PocketRocket 2 kamp ocağı', 1999.99, 40),
('550e8400-e29b-41d4-a716-446655440316', 'Big Agnes Copper Spur', 'Big Agnes Copper Spur HV UL2 çadır', 19999.99, 12),
('550e8400-e29b-41d4-a716-446655440317', 'Therm-a-Rest NeoAir', 'Therm-a-Rest NeoAir XLite mat', 6999.99, 25),
('550e8400-e29b-41d4-a716-446655440318', 'Kelty Cosmic 20', 'Kelty Cosmic 20 uyku tulumu', 4999.99, 30),
('550e8400-e29b-41d4-a716-446655440319', 'YOLO Board Inflatable SUP', 'YOLO Board şişirilebilir SUP tahtası', 14999.99, 10),
('550e8400-e29b-41d4-a716-446655440320', 'Wilson Pro Staff RF97', 'Wilson Pro Staff RF97 tenis raketi', 6999.99, 20),

-- Kitap, Müzik ve Sanat (401-500)
('550e8400-e29b-41d4-a716-446655440401', 'Atomic Habits - James Clear', 'Atomic Habits kitabı, alışkanlıklar üzerine', 199.99, 500),
('550e8400-e29b-41d4-a716-446655440402', 'Sapiens - Yuval Noah Harari', 'Sapiens: İnsanlığın Kısa Tarihi', 229.99, 300),
('550e8400-e29b-41d4-a716-446655440403', 'The 7 Habits - Stephen Covey', 'Etkili İnsanların 7 Alışkanlığı', 179.99, 250),
('550e8400-e29b-41d4-a716-446655440404', 'Rich Dad Poor Dad', 'Rich Dad Poor Dad - Robert Kiyosaki', 159.99, 400),
('550e8400-e29b-41d4-a716-446655440405', 'The Alchemist - Paulo Coelho', 'Simyacı - Paulo Coelho', 149.99, 350),
('550e8400-e29b-41d4-a716-446655440406', 'Harry Potter Seti', 'Harry Potter 7 kitap seti, Türkçe', 999.99, 100),
('550e8400-e29b-41d4-a716-446655440407', 'Game of Thrones Box Set', 'Game of Thrones 5 kitap seti, İngilizce', 1299.99, 75),
('550e8400-e29b-41d4-a716-446655440408', 'Kindle Paperwhite', 'Amazon Kindle Paperwhite e-kitap okuyucu', 3999.99, 80),
('550e8400-e29b-41d4-a716-446655440409', 'Kobo Clara 2E', 'Kobo Clara 2E e-kitap okuyucu', 3499.99, 60),
('550e8400-e29b-41d4-a716-446655440410', 'Yamaha P-45 Piyano', 'Yamaha P-45 dijital piyano, 88 tuş', 19999.99, 12),
('550e8400-e29b-41d4-a716-446655440411', 'Fender Player Stratocaster', 'Fender Player Stratocaster elektro gitar', 29999.99, 8),
('550e8400-e29b-41d4-a716-446655440412', 'Gibson Les Paul Standard', 'Gibson Les Paul Standard elektro gitar', 99999.99, 3),
('550e8400-e29b-41d4-a716-446655440413', 'Roland TD-17KV Drum Kit', 'Roland TD-17KV elektronik davul seti', 49999.99, 5),
('550e8400-e29b-41d4-a716-446655440414', 'Audio-Technica AT2020', 'Audio-Technica AT2020 kondenser mikrofon', 3999.99, 40),
('550e8400-e29b-41d4-a716-446655440415', 'Shure SM58', 'Shure SM58 dinamik vokal mikrofonu', 3499.99, 50),
('550e8400-e29b-41d4-a716-446655440416', 'Wacom Intuos Pro', 'Wacom Intuos Pro dijital çizim tableti', 12999.99, 25),
('550e8400-e29b-41d4-a716-446655440417', 'iPad Pro + Apple Pencil', 'iPad Pro 11" + Apple Pencil 2. nesil', 44999.99, 18),
('550e8400-e29b-41d4-a716-446655440418', 'Caran d''Ache Sanat Seti', 'Caran d''Ache profesyonel boyama seti', 2999.99, 30),
('550e8400-e29b-41d4-a716-446655440419', 'Moleskine Klasik Defter', 'Moleskine klasik defter, çizgili, büyük', 299.99, 200),
('550e8400-e29b-41d4-a716-446655440420', 'Parker Sonnet Dolma Kalem', 'Parker Sonnet altın kaplama dolma kalem', 1999.99, 50),
('550e8400-e29b-41d4-a716-446655440421', 'LEGO Architecture Taj Mahal', 'LEGO Architecture Taj Mahal seti', 1999.99, 25),
('550e8400-e29b-41d4-a716-446655440422', 'LEGO Technic Bugatti', 'LEGO Technic Bugatti Chiron seti', 12999.99, 8),
('550e8400-e29b-41d4-a716-446655440423', 'Ravensburger 5000 Puzzle', 'Ravensburger 5000 parça yetişkin puzzle', 799.99, 40),
('550e8400-e29b-41d4-a716-446655440424', 'Monopoly Deluxe', 'Monopoly Deluxe edition masa oyunu', 899.99, 60),
('550e8400-e29b-41d4-a716-446655440425', 'Chess.com Premium Set', 'Chess.com turnuva kalitesi satranç takımı', 2499.99, 30),
('550e8400-e29b-41d4-a716-446655440426', 'Magic: The Gathering Set', 'Magic: The Gathering başlangıç seti', 599.99, 100),
('550e8400-e29b-41d4-a716-446655440427', 'Pokemon Kart Seti', 'Pokemon TCG Battle Academy seti', 799.99, 80),
('550e8400-e29b-41d4-a716-446655440428', 'Uno Ultimate', 'Uno Ultimate kart oyunu', 299.99, 150),
('550e8400-e29b-41d4-a716-446655440429', 'Jenga Classic', 'Jenga Classic ahşap blok oyunu', 399.99, 120),
('550e8400-e29b-41d4-a716-446655440430', 'Twister Oyunu', 'Twister vücut eğlence oyunu', 449.99, 100),

-- Sağlık ve Güzellik (431-500)
('550e8400-e29b-41d4-a716-446655440431', 'Dyson Supersonic Saç Kurutma', 'Dyson Supersonic profesyonel saç kurutma makinesi', 14999.99, 20),
('550e8400-e29b-41d4-a716-446655440432', 'GHD Platinum+ Düzleştirici', 'GHD Platinum+ saç düzleştirici', 8999.99, 25),
('550e8400-e29b-41d4-a716-446655440433', 'Philips Norelco OneBlade', 'Philips Norelco OneBlade tıraş makinesi', 2999.99, 80),
('550e8400-e29b-41d4-a716-446655440434', 'Braun Series 9 Pro', 'Braun Series 9 Pro elektrikli tıraş makinesi', 12999.99, 30),
('550e8400-e29b-41d4-a716-446655440435', 'Oral-B Genius X', 'Oral-B Genius X akıllı diş fırçası', 4999.99, 50),
('550e8400-e29b-41d4-a716-446655440436', 'Waterpik Aquarius', 'Waterpik Aquarius su flosu', 2999.99, 40),
('550e8400-e29b-41d4-a716-446655440437', 'TheraGun Elite', 'TheraGun Elite masaj tabancası', 9999.99, 35),
('550e8400-e29b-41d4-a716-446655440438', 'NuFACE Trinity', 'NuFACE Trinity yüz tonlama cihazı', 11999.99, 15),
('550e8400-e29b-41d4-a716-446655440439', 'LED Yüz Maskesi', 'LED kırmızı ışık terapi yüz maskesi', 3999.99, 30),
('550e8400-e29b-41d4-a716-446655440440', 'Vitamin D3 Takviyesi', 'Vitamin D3 2000 IU günlük takviye', 199.99, 300),
('550e8400-e29b-41d4-a716-446655440441', 'Omega-3 Balık Yağı', 'Omega-3 balık yağı 1000mg kapsül', 299.99, 250),
('550e8400-e29b-41d4-a716-446655440442', 'Protein Tozu Whey', 'Whey protein tozu, çikolata aromalı', 899.99, 100),
('550e8400-e29b-41d4-a716-446655440443', 'Creatine Monohydrate', 'Creatine monohydrate toz, aromasız', 399.99, 150),
('550e8400-e29b-41d4-a716-446655440444', 'BCAA Aminoasit', 'BCAA aminoasit karışımı, limon aromalı', 599.99, 120),
('550e8400-e29b-41d4-a716-446655440445', 'Melatonin 3mg', 'Melatonin 3mg uyku takviyesi', 149.99, 200),
('550e8400-e29b-41d4-a716-446655440446', 'Magnesium Citrate', 'Magnesium citrate 400mg tablet', 179.99, 180),
('550e8400-e29b-41d4-a716-446655440447', 'Probiyotik Kapsül', 'Probiyotik 50 milyar CFU kapsül', 699.99, 100),
('550e8400-e29b-41d4-a716-446655440448', 'Collagen Peptides', 'Kollajen peptit toz, aromasız', 1299.99, 80),
('550e8400-e29b-41d4-a716-446655440449', 'Ashwagandha Extract', 'Ashwagandha ekstrakt 600mg kapsül', 499.99, 150),
('550e8400-e29b-41d4-a716-446655440450', 'Turmeric Curcumin', 'Turmeric curcumin 1000mg kapsül', 399.99, 120),

-- Otomotiv (451-500)
('550e8400-e29b-41d4-a716-446655440451', 'Michelin Pilot Sport 4S', 'Michelin Pilot Sport 4S lastik, 245/40R18', 4999.99, 40),
('550e8400-e29b-41d4-a716-446655440452', 'Bridgestone Potenza S001', 'Bridgestone Potenza S001 yaz lastiği', 3999.99, 50),
('550e8400-e29b-41d4-a716-446655440453', 'Continental WinterContact', 'Continental WinterContact kış lastiği', 3499.99, 60),
('550e8400-e29b-41d4-a716-446655440454', 'Bosch Icon Silecek', 'Bosch Icon silecek takımı, 24"+20"', 799.99, 100),
('550e8400-e29b-41d4-a716-446655440455', 'K&N Hava Filtresi', 'K&N yıkanabilir hava filtresi', 1499.99, 80),
('550e8400-e29b-41d4-a716-446655440456', 'Mobil 1 Synthetic Oil', 'Mobil 1 sentetik motor yağı 5W-30', 899.99, 120),
('550e8400-e29b-41d4-a716-446655440457', 'Castrol GTX Motor Oil', 'Castrol GTX motor yağı 10W-40', 599.99, 150),
('550e8400-e29b-41d4-a716-446655440458', 'NGK Spark Plugs', 'NGK iridium bujiler, 4lü set', 1299.99, 70),
('550e8400-e29b-41d4-a716-446655440459', 'Denso Spark Plugs', 'Denso platinum bujiler, 4lü set', 999.99, 90),
('550e8400-e29b-41d4-a716-446655440460', 'Thule Roof Box', 'Thule Motion XT XL tavan kutusu', 19999.99, 15),
('550e8400-e29b-41d4-a716-446655440461', 'Yakima SkyBox', 'Yakima SkyBox 21 tavan kutusu', 17999.99, 18),
('550e8400-e29b-41d4-a716-446655440462', 'WeatherTech FloorLiner', 'WeatherTech FloorLiner oto paspası', 2999.99, 50),
('550e8400-e29b-41d4-a716-446655440463', 'Husky Liners WeatherBeater', 'Husky Liners all-weather paspas', 1999.99, 60),
('550e8400-e29b-41d4-a716-446655440464', 'Chemical Guys Detailing Kit', 'Chemical Guys araç detaylandırma seti', 2499.99, 40),
('550e8400-e29b-41d4-a716-446655440465', 'Meguiars Ultimate Wax', 'Meguiars Ultimate araç cilası', 599.99, 80),
('550e8400-e29b-41d4-a716-446655440466', 'Armor All Car Cleaner', 'Armor All çok amaçlı araç temizleyici', 399.99, 100),
('550e8400-e29b-41d4-a716-446655440467', 'Garmin DriveSmart 66', 'Garmin DriveSmart 66 GPS navigasyon', 7999.99, 25),
('550e8400-e29b-41d4-a716-446655440468', 'TomTom GO 620', 'TomTom GO 620 akıllı navigasyon', 6999.99, 30),
('550e8400-e29b-41d4-a716-446655440469', 'Pioneer AVH-W4500NEX', 'Pioneer AVH-W4500NEX oto teyp', 12999.99, 20),
('550e8400-e29b-41d4-a716-446655440470', 'Alpine iLX-507', 'Alpine iLX-507 CarPlay stereo', 9999.99, 25),
('550e8400-e29b-41d4-a716-446655440471', 'JL Audio Subwoofer', 'JL Audio 12" aktif subwoofer', 14999.99, 15),
('550e8400-e29b-41d4-a716-446655440472', 'Kicker CompVR', 'Kicker CompVR 12" subwoofer', 7999.99, 20),
('550e8400-e29b-41d4-a716-446655440473', 'Dashcam 4K', '4K Ultra HD araç kamerası, GPS', 4999.99, 40),
('550e8400-e29b-41d4-a716-446655440474', 'BlackVue DR900X', 'BlackVue DR900X 4K dashcam', 19999.99, 12),
('550e8400-e29b-41d4-a716-446655440475', 'NOCO Boost Plus GB40', 'NOCO Boost Plus GB40 akü takviye', 3999.99, 35),
('550e8400-e29b-41d4-a716-446655440476', 'CTEK MXS 5.0', 'CTEK MXS 5.0 akü şarj cihazı', 2999.99, 30),
('550e8400-e29b-41d4-a716-446655440477', 'Anker Roav Jump Starter', 'Anker Roav akü takviye + powerbank', 2499.99, 45),
('550e8400-e29b-41d4-a716-446655440478', 'Cobra RAD 480i', 'Cobra RAD 480i radar dedektörü', 5999.99, 20),
('550e8400-e29b-41d4-a716-446655440479', 'Escort MAX 360c', 'Escort MAX 360c WiFi radar dedektörü', 19999.99, 8),
('550e8400-e29b-41d4-a716-446655440480', 'Ring Automotive Emergency Kit', 'Ring acil durum araç seti', 1999.99, 50),

-- Bonus Kategoriler (481-500)
('550e8400-e29b-41d4-a716-446655440481', 'Tesla Cybertruck Model', 'Tesla Cybertruck 1:24 ölçek model', 799.99, 100),
('550e8400-e29b-41d4-a716-446655440482', 'SpaceX Falcon 9 Model', 'SpaceX Falcon 9 roket modeli', 1299.99, 50),
('550e8400-e29b-41d4-a716-446655440483', 'Star Wars LEGO Millennium', 'LEGO Star Wars Millennium Falcon', 29999.99, 5),
('550e8400-e29b-41d4-a716-446655440484', 'Marvel Action Figures Set', 'Marvel Avengers aksiyon figür seti', 2999.99, 30),
('550e8400-e29b-41d4-a716-446655440485', 'Pokemon Plush Collection', 'Pokemon peluş oyuncak koleksiyonu', 1999.99, 60),
('550e8400-e29b-41d4-a716-446655440486', 'Funko Pop Stranger Things', 'Funko Pop Stranger Things karakter seti', 899.99, 80),
('550e8400-e29b-41d4-a716-446655440487', 'Hot Wheels Track Builder', 'Hot Wheels Track Builder mega set', 1499.99, 40),
('550e8400-e29b-41d4-a716-446655440488', 'Nerf Elite 2.0 Commander', 'Nerf Elite 2.0 Commander blaster', 1299.99, 70),
('550e8400-e29b-41d4-a716-446655440489', 'LEGO Creator Expert Car', 'LEGO Creator Expert Ford Mustang', 5999.99, 15),
('550e8400-e29b-41d4-a716-446655440490', 'Remote Control Drone', 'Profesyonel kamera drone, 4K video', 12999.99, 20),
('550e8400-e29b-41d4-a716-446655440491', 'RC Car High Speed', 'RC yüksek hızlı araba, 50km/h', 3999.99, 25),
('550e8400-e29b-41d4-a716-446655440492', 'Electric Scooter', 'Elektrikli scooter, 25km menzil', 8999.99, 15),
('550e8400-e29b-41d4-a716-446655440493', 'Hoverboard Self Balance', 'Hoverboard self-balancing scooter', 4999.99, 20),
('550e8400-e29b-41d4-a716-446655440494', 'Electric Skateboard', 'Elektrikli kaykay, uzaktan kumanda', 14999.99, 10),
('550e8400-e29b-41d4-a716-446655440495', 'VR Headset Meta Quest 2', 'Meta Quest 2 VR başlığı, 256GB', 12999.99, 25),
('550e8400-e29b-41d4-a716-446655440496', 'PlayStation VR2', 'PlayStation VR2 sanal gerçeklik seti', 21999.99, 12),
('550e8400-e29b-41d4-a716-446655440497', 'Nintendo Labo VR Kit', 'Nintendo Labo VR Kit Switch için', 2999.99, 30),
('550e8400-e29b-41d4-a716-446655440498', 'Raspberry Pi 4 Kit', 'Raspberry Pi 4 başlangıç seti', 2499.99, 40),
('550e8400-e29b-41d4-a716-446655440499', 'Arduino Uno Starter Kit', 'Arduino Uno başlangıç seti', 1499.99, 60),
('550e8400-e29b-41d4-a716-446655440500', 'Micro:bit Education Set', 'BBC micro:bit eğitim seti', 1999.99, 50)
ON CONFLICT DO NOTHING;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_customer_email ON orders(customer_email);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_order_id ON payment_transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_event_type ON order_status_history(event_type);
CREATE INDEX IF NOT EXISTS idx_shipments_order_id ON shipments(order_id);
CREATE INDEX IF NOT EXISTS idx_shipments_tracking_number ON shipments(tracking_number);
CREATE INDEX IF NOT EXISTS idx_stock_reservations_product_id ON stock_reservations(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_reservations_order_id ON stock_reservations(order_id);

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
CREATE TRIGGER update_orders_updated_at 
    BEFORE UPDATE ON orders 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
CREATE TRIGGER update_payments_updated_at 
    BEFORE UPDATE ON payments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_payment_transactions_updated_at ON payment_transactions;
CREATE TRIGGER update_payment_transactions_updated_at 
    BEFORE UPDATE ON payment_transactions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_shipments_updated_at ON shipments;
CREATE TRIGGER update_shipments_updated_at 
    BEFORE UPDATE ON shipments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Grant permissions to orderuser (if needed)
-- This is automatically handled by PostgreSQL when using POSTGRES_USER in Docker 