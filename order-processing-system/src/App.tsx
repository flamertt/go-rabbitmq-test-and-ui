import { useState, useEffect } from 'react';
import './App.css';

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

const API_BASE_URL = 'http://localhost:8080/api/v1';

function App() {
  const [products, setProducts] = useState<Product[]>([]);
  const [cart, setCart] = useState<CartItem[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [activeTab, setActiveTab] = useState<'products' | 'cart' | 'orders'>('products');
  const [notification, setNotification] = useState<string>('');

  useEffect(() => {
    fetchProducts();
    fetchOrders();
  }, []);

  const fetchProducts = async () => {
    try {
      setLoading(true);
      const response = await fetch(`${API_BASE_URL}/products`);
      if (response.ok) {
        const data = await response.json();
        setProducts(data);
      }
    } catch (error) {
      showNotification('Ürünler yüklenirken hata oluştu');
    } finally {
      setLoading(false);
    }
  };

  const fetchOrders = async () => {
    try {
      const storedOrders = localStorage.getItem('orders');
      if (storedOrders) {
        setOrders(JSON.parse(storedOrders));
      }
    } catch (error) {
      console.error('Siparişler yüklenirken hata:', error);
    }
  };

  const addToCart = (product: Product) => {
    setCart(prevCart => {
      const existingItem = prevCart.find(item => item.id === product.id);
      if (existingItem) {
        return prevCart.map(item =>
          item.id === product.id
            ? { ...item, quantity: Math.min(item.quantity + 1, product.stock_quantity) }
            : item
        );
      }
      return [...prevCart, { ...product, quantity: 1 }];
    });
    showNotification(`${product.name} sepete eklendi`);
  };

  const removeFromCart = (productId: string) => {
    setCart(prevCart => prevCart.filter(item => item.id !== productId));
  };

  const updateCartQuantity = (productId: string, quantity: number) => {
    if (quantity <= 0) {
      removeFromCart(productId);
      return;
    }
    
    setCart(prevCart =>
      prevCart.map(item =>
        item.id === productId ? { ...item, quantity } : item
      )
    );
  };

  const calculateTotal = () => {
    return cart.reduce((total, item) => total + (item.price * item.quantity), 0);
  };

  const createOrder = async () => {
    if (cart.length === 0) {
      showNotification('Sepetiniz boş');
      return;
    }

    try {
      setLoading(true);
      const orderData = {
        user_id: 'demo-user-' + Date.now(),
        items: cart.map(item => ({
          product_id: item.id,
          quantity: item.quantity
        }))
      };

      const response = await fetch(`${API_BASE_URL}/orders`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(orderData)
      });

      if (response.ok) {
        const order = await response.json();
        const newOrders = [...orders, order];
        setOrders(newOrders);
        localStorage.setItem('orders', JSON.stringify(newOrders));
        setCart([]);
        setActiveTab('orders');
        showNotification('Sipariş başarıyla oluşturuldu!');
        fetchProducts(); // Refresh products to update stock
      } else {
        const error = await response.json();
        showNotification(error.error || 'Sipariş oluşturulurken hata oluştu');
      }
    } catch (error) {
      showNotification('Bağlantı hatası oluştu');
    } finally {
      setLoading(false);
    }
  };

  const showNotification = (message: string) => {
    setNotification(message);
    setTimeout(() => setNotification(''), 3000);
  };

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
              <h2 className="section-title">Ürünler</h2>
              <div className="products-grid">
                {products.map(product => (
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
                ))}
              </div>
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
                      onClick={createOrder}
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
                    <div key={order.order_id} className="order-card">
                      <div className="order-header">
                        <div className="order-id">#{order.order_id.slice(-8)}</div>
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

export default App;
