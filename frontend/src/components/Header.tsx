export function Header({ store }: { store: any }) {
  return (
    <header className="glass-panel header animate-fade-in">
      <div className="container header-content">
        <div className="logo-section">
          <div className="logo-mark">⚡</div>
          <h1>Store</h1>
        </div>
        
        <div className="nav-tabs">
          <button 
            className={`tab-btn ${store.activeTab === 'shop' ? 'active' : ''}`}
            onClick={() => store.setActiveTab('shop')}
          >
            Shop
          </button>
          <button 
            className={`tab-btn ${store.activeTab === 'inventory' ? 'active' : ''}`}
            onClick={() => store.setActiveTab('inventory')}
          >
            Inventory
          </button>
          <button 
            className={`tab-btn ${store.activeTab === 'orders' ? 'active' : ''}`}
            onClick={() => store.setActiveTab('orders')}
          >
            Order History
          </button>
        </div>

        <div className="actions-section">
          {store.activeTab === 'shop' && (
            <input 
              type="text" 
              className="search-input"
              placeholder="Search catalog..." 
              value={store.search}
              onChange={(e) => store.setSearch(e.target.value)}
            />
          )}

          {store.activeTab === 'inventory' && (
            <>
              <input 
                type="text" 
                className="search-input"
                placeholder="Search inventory..." 
                value={store.search}
                onChange={(e) => store.setSearch(e.target.value)}
              />
              <button 
                className="btn-primary"
                onClick={() => store.openModal()}
              >
                + Add Product
              </button>
              <button 
                className="btn-outline upload-btn"
                onClick={() => store.fileInputRef.current?.click()}
                disabled={store.isImporting}
              >
                {store.isImporting ? 'Importing...' : 'Upload CSV'}
              </button>
              <input 
                type="file" 
                accept=".csv" 
                ref={store.fileInputRef} 
                style={{ display: 'none' }} 
                onChange={store.handleFileUpload}
              />
            </>
          )}
        </div>
      </div>
    </header>
  );
}
