import { useState, useEffect } from 'react';
import './App.css';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import AuthForm from './components/AuthForm';

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  stock_quantity: number;
}

interface CartItem extends Product {
  quantity: number;
}

interface Order {
  order_id: string;
  user_id: string;
  total_amount: number;
  status: string;
  message: string;
  created_at?: string;
}

interface PaginatedResponse {
  data: Product[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

interface ProductFilters {
  search: string;
  category: string;
  minPrice: string;
  maxPrice: string;
  sortBy: string;
  sortDir: string;
}

const API_BASE_URL = 'http://localhost:8080/api/v1';

function MainApp() {
  const { user, logout, isAuthenticated, token } = useAuth();
  const [products, setProducts] = useState<Product[]>([]);
  const [cart, setCart] = useState<CartItem[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [activeTab, setActiveTab] = useState<'products' | 'cart' | 'orders'>('products');
  const [notification, setNotification] = useState<string>('');
  
  // Pagination states
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [pageSize] = useState<number>(20);
  const [totalPages, setTotalPages] = useState<number>(1);
  const [totalProducts, setTotalProducts] = useState<number>(0);
  
  // Filter states
  const [filters, setFilters] = useState<ProductFilters>({
    search: '',
    category: '',
    minPrice: '',
    maxPrice: '',
    sortBy: 'created_at',
    sortDir: 'desc'
  });

  useEffect(() => {
    if (isAuthenticated) {
      fetchProducts();
      fetchOrders();
    }
  }, [isAuthenticated, currentPage, filters]);

  const getAuthHeaders = () => {
    return {
      'Content-Type': 'application/json',
      'Authorization': token ? `Bearer ${token}` : '',
    };
  };

  const buildProductsUrl = () => {
    const params = new URLSearchParams({
      page: currentPage.toString(),
      page_size: pageSize.toString(),
      sort_by: filters.sortBy,
      sort_dir: filters.sortDir
    });

    if (filters.search) params.append('search', filters.search);
    if (filters.category) params.append('category', filters.category);
    if (filters.minPrice) params.append('min_price', filters.minPrice);
    if (filters.maxPrice) params.append('max_price', filters.maxPrice);

    return `${API_BASE_URL}/products?${params.toString()}`;
  };

  const fetchProducts = async () => {
    try {
      setLoading(true);
      const response = await fetch(buildProductsUrl(), {
        headers: getAuthHeaders(),
      });
      if (response.ok) {
        const data: PaginatedResponse = await response.json();
        setProducts(data.data || []);
        setTotalPages(data.total_pages || 1);
        setTotalProducts(data.total || 0);
      } else if (response.status === 401) {
        logout();
      } else {
        showNotification('Ürünler yüklenirken hata oluştu');
        setProducts([]);
        setTotalPages(1);
        setTotalProducts(0);
      }
    } catch (error) {
      console.error('Products fetch error:', error);
      showNotification('Ürünler yüklenirken hata oluştu');
      setProducts([]);
      setTotalPages(1);
      setTotalProducts(0);
    } finally {
      setLoading(false);
    }
  };

  const fetchOrders = async () => {
    if (!user?.id) {
      console.error('User ID not available');
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/orders?user_id=${user.id}`, {
        headers: getAuthHeaders(),
      });
      if (response.ok) {
        const data = await response.json();
        setOrders(data || []);
      } else if (response.status === 401) {
        logout();
      }
    } catch (error) {
      console.error('Siparişler yüklenirken hata:', error);
    }
  };

  const addToCart = (product: Product) => {
    setCart(currentCart => {
      const existingItem = currentCart.find(item => item.id === product.id);
      if (existingItem) {
        return currentCart.map(item =>
          item.id === product.id
            ? { ...item, quantity: item.quantity + 1 }
            : item
        );
      }
      return [...currentCart, { ...product, quantity: 1 }];
    });
    showNotification(`${product.name} sepete eklendi!`);
  };

  const removeFromCart = (productId: string) => {
    setCart(currentCart => currentCart.filter(item => item.id !== productId));
    showNotification('Ürün sepetten kaldırıldı');
  };

  const updateCartQuantity = (productId: string, quantity: number) => {
    if (quantity <= 0) {
      removeFromCart(productId);
      return;
    }

    setCart(currentCart =>
      currentCart.map(item =>
        item.id === productId ? { ...item, quantity } : item
      )
    );
  };

  const calculateTotal = () => {
    return cart.reduce((total, item) => total + (item.price * item.quantity), 0);
  };

  const handleOrderSubmit = async () => {
    if (!user) {
      showNotification('Sipariş vermek için giriş yapmalısınız');
      return;
    }

    if (cart.length === 0) {
      showNotification('Sepetiniz boş!');
      return;
    }

    try {
      setLoading(true);
      const orderData = {
        user_id: user.id,
        customer_email: user.email,
        items: cart.map(item => ({
          product_id: item.id,
          quantity: item.quantity,
          price: item.price
        })),
        total_amount: calculateTotal()
      };

      const response = await fetch(`${API_BASE_URL}/orders`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(orderData),
      });

      if (response.ok) {
        const result = await response.json();
        showNotification(`Sipariş başarıyla oluşturuldu! Sipariş ID: ${result.order_id}`);
        setCart([]);
        setActiveTab('orders');
        await fetchOrders();
      } else if (response.status === 401) {
        logout();
      } else {
        const errorData = await response.json();
        showNotification(`Sipariş oluşturulamadı: ${errorData.message || 'Bilinmeyen hata'}`);
      }
    } catch (error) {
      showNotification('Sipariş gönderilirken hata oluştu');
      console.error('Order submission error:', error);
    } finally {
      setLoading(false);
    }
  };

  const showNotification = (message: string) => {
    setNotification(message);
    setTimeout(() => setNotification(''), 3000);
  };

  const handleLogout = () => {
    logout();
    setCart([]);
    setOrders([]);
    setProducts([]);
  };

  if (!isAuthenticated) {
    return <AuthForm />;
  }

  const getStatusColor = (status: string) => {
    const statusColors: { [key: string]: string } = {
      'CREATED': '#3498db',
      'PAYMENT_SUCCESSFUL': '#2ecc71',
      'PAYMENT_FAILED': '#e74c3c',
      'STOCK_RESERVED': '#f39c12',
      'STOCK_INSUFFICIENT': '#e74c3c',
      'SHIPPED': '#9b59b6',
      'DELIVERED': '#27ae60',
      'CANCELLED': '#95a5a6'
    };
    return statusColors[status] || '#95a5a6';
  };

  return (
    <div className="app">
      {notification && (
        <div className="notification">
          {notification}
        </div>
      )}

      <header className="header">
        <div className="container">
          <h1 className="logo">Order System</h1>
          <nav className="nav">
            <button 
              className={`nav-button ${activeTab === 'products' ? 'active' : ''}`}
              onClick={() => setActiveTab('products')}
            >
              Ürünler ({products.length})
            </button>
            <button 
              className={`nav-button ${activeTab === 'cart' ? 'active' : ''}`}
              onClick={() => setActiveTab('cart')}
            >
              Sepet ({cart.length})
            </button>
            <button 
              className={`nav-button ${activeTab === 'orders' ? 'active' : ''}`}
              onClick={() => setActiveTab('orders')}
            >
              Siparişlerim ({orders.length})
            </button>
          </nav>
          <div className="user-info">
            <span className="user-name">
              Hoş geldin, {user?.first_name} {user?.last_name}
            </span>
            <button 
              className="logout-button"
              onClick={handleLogout}
            >
              Çıkış
            </button>
          </div>
        </div>
      </header>

      <main className="main">
        <div className="container">
          {loading && (
            <div className="loading">
              <div className="spinner"></div>
              <p>Yükleniyor...</p>
            </div>
          )}

          {activeTab === 'products' && (
            <section className="section">
              <div className="products-header">
                <h2 className="section-title">Ürünler</h2>
                <div className="products-stats">
                  <span>{totalProducts} ürün bulundu</span>
                  <span>Sayfa {currentPage} / {totalPages}</span>
                </div>
              </div>

              {/* Filter and Search Controls */}
              <div className="products-filters">
                <div className="search-box">
                  <input
                    type="text"
                    placeholder="Ürün adı veya açıklama ara..."
                    value={filters.search}
                    onChange={(e) => {
                      setFilters(prev => ({ ...prev, search: e.target.value }));
                      setCurrentPage(1);
                    }}
                    className="search-input"
                  />
                </div>
                
                <div className="filter-controls">
                  <div className="price-filters">
                    <input
                      type="number"
                      placeholder="Min fiyat"
                      value={filters.minPrice}
                      onChange={(e) => {
                        setFilters(prev => ({ ...prev, minPrice: e.target.value }));
                        setCurrentPage(1);
                      }}
                      className="price-input"
                    />
                    <input
                      type="number"
                      placeholder="Max fiyat"
                      value={filters.maxPrice}
                      onChange={(e) => {
                        setFilters(prev => ({ ...prev, maxPrice: e.target.value }));
                        setCurrentPage(1);
                      }}
                      className="price-input"
                    />
                  </div>

                  <select
                    value={`${filters.sortBy}-${filters.sortDir}`}
                    onChange={(e) => {
                      const [sortBy, sortDir] = e.target.value.split('-');
                      setFilters(prev => ({ ...prev, sortBy, sortDir }));
                    }}
                    className="sort-select"
                  >
                    <option value="created_at-desc">En Yeni</option>
                    <option value="created_at-asc">En Eski</option>
                    <option value="price-asc">Fiyat: Düşük → Yüksek</option>
                    <option value="price-desc">Fiyat: Yüksek → Düşük</option>
                    <option value="name-asc">İsim: A → Z</option>
                    <option value="name-desc">İsim: Z → A</option>
                    <option value="stock-desc">Stok: Çok → Az</option>
                    <option value="stock-asc">Stok: Az → Çok</option>
                  </select>

                  <button
                    onClick={() => {
                      setFilters({
                        search: '',
                        category: '',
                        minPrice: '',
                        maxPrice: '',
                        sortBy: 'created_at',
                        sortDir: 'desc'
                      });
                      setCurrentPage(1);
                    }}
                    className="clear-filters-btn"
                  >
                    Filtreleri Temizle
                  </button>
                </div>
              </div>

              <div className="products-grid">
                {products && products.length > 0 ? products.map(product => (
                  <div key={product.id} className="product-card">
                    <div className="product-header">
                      <h3 className="product-name">{product.name}</h3>
                      <span className="product-price">₺{product.price.toFixed(2)}</span>
                    </div>
                    <p className="product-description">{product.description}</p>
                    <div className="product-footer">
                      <span className={`stock-badge ${product.stock_quantity > 10 ? 'in-stock' : product.stock_quantity > 0 ? 'low-stock' : 'out-of-stock'}`}>
                        {product.stock_quantity > 0 ? `${product.stock_quantity} adet` : 'Stokta yok'}
                      </span>
                      <button 
                        className="add-to-cart-btn"
                        onClick={() => addToCart(product)}
                        disabled={product.stock_quantity === 0}
                      >
                        Sepete Ekle
                      </button>
                    </div>
                  </div>
                )) : (
                  <div className="empty-state">
                    <p>Hiç ürün bulunamadı</p>
                    {loading && <p>Yükleniyor...</p>}
                  </div>
                )}
              </div>

              {/* Pagination Controls */}
              {totalPages && totalPages > 1 && (
                <div className="pagination">
                  <button
                    onClick={() => setCurrentPage(1)}
                    disabled={currentPage === 1}
                    className="pagination-btn"
                  >
                    ⏮ İlk
                  </button>
                  <button
                    onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                    disabled={currentPage === 1}
                    className="pagination-btn"
                  >
                    ⬅ Önceki
                  </button>
                  
                  <div className="pagination-info">
                    {Math.max(1, currentPage - 2) !== 1 && (
                      <>
                        <button onClick={() => setCurrentPage(1)} className="pagination-number">1</button>
                        {Math.max(1, currentPage - 2) > 2 && <span className="pagination-dots">...</span>}
                      </>
                    )}
                    
                    {Array.from(
                      { length: Math.min(5, totalPages || 1) },
                      (_, i) => Math.max(1, currentPage - 2) + i
                    )
                      .filter(page => page <= (totalPages || 1))
                      .map(page => (
                        <button
                          key={page}
                          onClick={() => setCurrentPage(page)}
                          className={`pagination-number ${page === currentPage ? 'active' : ''}`}
                        >
                          {page}
                        </button>
                      ))}
                    
                    {Math.min(totalPages || 1, currentPage + 2) !== (totalPages || 1) && (totalPages || 1) > 5 && (
                      <>
                        {Math.min(totalPages || 1, currentPage + 2) < (totalPages || 1) - 1 && <span className="pagination-dots">...</span>}
                        <button onClick={() => setCurrentPage(totalPages || 1)} className="pagination-number">{totalPages}</button>
                      </>
                    )}
                  </div>

                  <button
                    onClick={() => setCurrentPage(prev => Math.min(totalPages || 1, prev + 1))}
                    disabled={currentPage === (totalPages || 1)}
                    className="pagination-btn"
                  >
                    Sonraki ➡
                  </button>
                  <button
                    onClick={() => setCurrentPage(totalPages || 1)}
                    disabled={currentPage === (totalPages || 1)}
                    className="pagination-btn"
                  >
                    Son ⏭
                  </button>
                </div>
              )}
            </section>
          )}

          {activeTab === 'cart' && (
            <section className="section">
              <h2 className="section-title">Sepetim</h2>
              {cart.length === 0 ? (
                <div className="empty-state">
                  <p>Sepetiniz boş</p>
                  <button 
                    className="primary-button"
                    onClick={() => setActiveTab('products')}
                  >
                    Alışverişe Başla
                  </button>
                </div>
              ) : (
                <div className="cart-container">
                  <div className="cart-items">
                    {cart.map(item => (
                      <div key={item.id} className="cart-item">
                        <div className="cart-item-info">
                          <h4 className="cart-item-name">{item.name}</h4>
                          <p className="cart-item-price">₺{item.price.toFixed(2)}</p>
                        </div>
                        <div className="cart-item-controls">
                          <div className="quantity-controls">
                            <button 
                              className="quantity-btn"
                              onClick={() => updateCartQuantity(item.id, item.quantity - 1)}
                            >
                              -
                            </button>
                            <span className="quantity">{item.quantity}</span>
                            <button 
                              className="quantity-btn"
                              onClick={() => updateCartQuantity(item.id, item.quantity + 1)}
                              disabled={item.quantity >= item.stock_quantity}
                            >
                              +
                            </button>
                          </div>
                          <button 
                            className="remove-btn"
                            onClick={() => removeFromCart(item.id)}
                          >
                            Kaldır
                          </button>
                        </div>
                        <div className="cart-item-total">
                          ₺{(item.price * item.quantity).toFixed(2)}
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="cart-summary">
                    <div className="total-row">
                      <span className="total-label">Toplam:</span>
                      <span className="total-amount">₺{calculateTotal().toFixed(2)}</span>
                    </div>
                    <button 
                      className="checkout-btn"
                      onClick={handleOrderSubmit}
                      disabled={loading}
                    >
                      Siparişi Tamamla
                    </button>
                  </div>
                </div>
              )}
            </section>
          )}

          {activeTab === 'orders' && (
            <section className="section">
              <h2 className="section-title">Siparişlerim</h2>
              {orders.length === 0 ? (
                <div className="empty-state">
                  <p>Henüz siparişiniz bulunmuyor</p>
                  <button 
                    className="primary-button"
                    onClick={() => setActiveTab('products')}
                  >
                    Alışverişe Başla
                  </button>
                </div>
              ) : (
                <div className="orders-list">
                  {orders.map(order => (
                    <div key={order.order_id || 'unknown'} className="order-card">
                      <div className="order-header">
                        <div className="order-id">#{order.order_id ? order.order_id.slice(-8) : 'UNKNOWN'}</div>
                        <div 
                          className="order-status"
                          style={{ backgroundColor: getStatusColor(order.status) }}
                        >
                          {order.status}
                        </div>
                      </div>
                      <div className="order-details">
                        <div className="order-amount">₺{order.total_amount.toFixed(2)}</div>
                        <div className="order-message">{order.message}</div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>
          )}
        </div>
      </main>

      <footer className="footer">
        <div className="container">
          <p>&copy; 2025 Order Processing System.</p>
        </div>
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <MainApp />
    </AuthProvider>
  );
}
