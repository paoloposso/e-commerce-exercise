import { useAppStore } from './hooks/useAppStore';
import { Header } from './components/Header';
import { Catalog } from './components/Catalog';
import { OrdersTable } from './components/OrdersTable';
import { ProductModal } from './components/ProductModal';
import './App.css';

function App() {
  const store = useAppStore();

  return (
    <div className="app-container">
      <Header store={store} />

      <main className="container main-content">
        {store.toast && (
          <div className={`toast toast-${store.toast.type} animate-fade-in`}>
            {store.toast.message}
          </div>
        )}

        {store.loading ? (
          <div className="loading-state">Loading data...</div>
        ) : store.error ? (
          <div className="error-state">{store.error}</div>
        ) : store.activeTab === 'shop' ? (
          <Catalog store={store} mode="shop" />
        ) : store.activeTab === 'inventory' ? (
          <Catalog store={store} mode="admin" />
        ) : (
          <OrdersTable store={store} />
        )}
      </main>

      <ProductModal store={store} />
    </div>
  );
}

export default App;
