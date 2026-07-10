import { ProductCard } from './ProductCard';
import type { Product } from '../services/api';

export function Catalog({ store, mode }: { store: any, mode: 'shop' | 'admin' }) {
  if (store.products.length === 0) {
    return (
      <div className="empty-state glass-panel">
        <h2>No products found</h2>
        <p>{mode === 'admin' ? 'Add products manually or upload a CSV file.' : 'Upload a CSV file to populate the catalog.'}</p>
      </div>
    );
  }

  return (
    <div className="product-grid">
      {store.products.map((p: Product, i: number) => (
        <ProductCard key={p.id} p={p} index={i} store={store} mode={mode} />
      ))}
    </div>
  );
}
