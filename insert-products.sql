-- Insert products in batches to avoid PostgreSQL statement size limits

-- Batch 1: Electronics (25 products)
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
('550e8400-e29b-41d4-a716-446655440025', 'AMD Ryzen 9 7950X', 'AMD Ryzen 9 7950X işlemci, 16 çekirdek', 24999.99, 15)
ON CONFLICT (id) DO NOTHING;

-- Batch 2: Clothing & Fashion (25 products)  
INSERT INTO products (id, name, description, price, stock_quantity) VALUES
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
('550e8400-e29b-41d4-a716-446655440121', 'North Face Jacket', 'The North Face outdoor mont, su geçirmez', 3499.99, 40),
('550e8400-e29b-41d4-a716-446655440122', 'Columbia Fleece', 'Columbia polar fleece ceket', 1999.99, 65),
('550e8400-e29b-41d4-a716-446655440123', 'Patagonia T-Shirt', 'Patagonia organik pamuk t-shirt', 899.99, 110),
('550e8400-e29b-41d4-a716-446655440124', 'Lululemon Leggings', 'Lululemon yoga tayt, nefes alabilir kumaş', 3999.99, 80),
('550e8400-e29b-41d4-a716-446655440125', 'Champion Sweatshirt', 'Champion vintage logo sweatshirt', 1299.99, 100)
ON CONFLICT (id) DO NOTHING; 