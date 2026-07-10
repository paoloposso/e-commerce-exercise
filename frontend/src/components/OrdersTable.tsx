import type { PurchaseResult } from '../services/api';

export function OrdersTable({ store }: { store: any }) {
  if (store.orders.length === 0) {
    return (
      <div className="empty-state glass-panel">
        <h2>No orders yet</h2>
        <p>Go to the catalog and purchase something!</p>
      </div>
    );
  }

  return (
    <div className="orders-table-container glass-panel animate-fade-in">
      <table className="orders-table">
        <thead>
          <tr>
            <th>Order ID</th>
            <th>Date</th>
            <th>Customer</th>
            <th>SKU</th>
            <th>Qty</th>
            <th>Total Price</th>
          </tr>
        </thead>
        <tbody>
          {store.orders.map((o: PurchaseResult) => (
            <tr key={o.id}>
              <td className="monospace">{o.id.substring(0, 8)}...</td>
              <td>{o.created_at ? new Date(o.created_at).toLocaleString() : '-'}</td>
              <td>{o.customer_id}</td>
              <td className="monospace">{o.sku}</td>
              <td>{o.quantity}</td>
              <td className="success-text">${o.total_price.toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
