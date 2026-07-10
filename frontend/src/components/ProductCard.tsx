import type { Product } from '../services/api';

export function ProductCard({ p, index, store, mode }: { p: Product, index: number, store: any, mode: 'shop' | 'admin' }) {
  const qty = store.quantities[p.sku] || 1;

  return (
    <div 
      className="product-card glass-panel animate-fade-in"
      style={{ animationDelay: `${index * 0.05}s` }}
    >
      {mode === 'admin' && (
        <div className="product-actions">
          <button 
            className="btn-icon" 
            onClick={() => store.openModal(p)}
            title="Edit Product"
          >
            ✏️
          </button>
          <button 
            className="btn-icon btn-icon-delete" 
            onClick={() => store.handleDeleteProduct(p)}
            title="Delete Product"
          >
            🗑️
          </button>
        </div>
      )}

      <div className="product-category">{p.category || 'General'}</div>
      <h3 className="product-title">{p.name}</h3>
      <p className="product-sku">SKU: {p.sku}</p>
      <p className="product-desc">{p.description}</p>
      
      <div className="product-footer">
        <div className="price-stock">
          <span className="price">
            ${mode === 'shop' ? (p.price * qty).toFixed(2) : p.price.toFixed(2)}
          </span>
          <span className={`stock ${p.stock > 0 ? 'in-stock' : 'out-of-stock'}`}>
            {p.stock > 0 ? `${p.stock} in stock` : 'Out of stock'}
          </span>
        </div>
        
        {mode === 'shop' && (
          <div className="purchase-actions">
            <div className="quantity-selector">
              <button 
                className="qty-btn" 
                disabled={p.stock <= 0 || qty <= 1}
                onClick={() => store.handleQuantityChange(p.sku, -1, p.stock)}
              >-</button>
              <span className="qty-display">{p.stock <= 0 ? 0 : qty}</span>
              <button 
                className="qty-btn" 
                disabled={p.stock <= 0 || qty >= p.stock}
                onClick={() => store.handleQuantityChange(p.sku, 1, p.stock)}
              >+</button>
            </div>
            <button 
              className="btn-primary" 
              disabled={p.stock <= 0}
              onClick={() => store.handlePurchase(p)}
            >
              Buy Now
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
