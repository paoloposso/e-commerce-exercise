export function ProductModal({ store }: { store: any }) {
  if (!store.isModalOpen) return null;

  return (
    <div className="modal-backdrop" onClick={() => store.setIsModalOpen(false)}>
      <div className="modal-content glass-panel" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>{store.editingProduct ? 'Edit Product' : 'Add New Product'}</h2>
          <button className="btn-close" onClick={() => store.setIsModalOpen(false)}>&times;</button>
        </div>
        <form onSubmit={store.handleSaveProduct} noValidate>
          <div className="form-group">
            <label htmlFor="name">Product Name *</label>
            <input 
              type="text" 
              id="name"
              className="form-input" 
              value={store.formName} 
              onChange={(e) => store.setFormName(e.target.value)} 
              required 
            />
            {store.fieldErrors.name && (
              <span className="field-error">{store.fieldErrors.name}</span>
            )}
          </div>
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="sku">SKU *</label>
              <input 
                type="text" 
                id="sku"
                className="form-input" 
                value={store.formSku} 
                onChange={(e) => store.setFormSku(e.target.value)} 
                required 
              />
              {store.fieldErrors.sku && (
                <span className="field-error">{store.fieldErrors.sku}</span>
              )}
            </div>
            <div className="form-group">
              <label htmlFor="category">Category</label>
              <input 
                type="text" 
                id="category"
                className="form-input" 
                value={store.formCategory} 
                onChange={(e) => store.setFormCategory(e.target.value)} 
              />
            </div>
          </div>
          <div className="form-group">
            <label htmlFor="description">Description</label>
            <textarea 
              id="description"
              className="form-textarea" 
              rows={3} 
              value={store.formDescription} 
              onChange={(e) => store.setFormDescription(e.target.value)} 
            />
          </div>
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="price">Price ($) *</label>
              <input 
                type="text" 
                id="price"
                className="form-input" 
                inputMode="numeric"
                value={store.formPrice.toFixed(2)} 
                onChange={(e) => store.handlePriceChange(e.target.value)} 
                required 
              />
              {store.fieldErrors.price && (
                <span className="field-error">{store.fieldErrors.price}</span>
              )}
            </div>
            <div className="form-group">
              <label htmlFor="stock">Stock *</label>
              <input 
                type="number" 
                id="stock"
                className="form-input" 
                min="0"
                value={store.formStock} 
                onChange={(e) => store.setFormStock(Number(e.target.value))} 
                required 
              />
              {store.fieldErrors.stock && (
                <span className="field-error">{store.fieldErrors.stock}</span>
              )}
            </div>
          </div>
          <div className="form-group">
            <label htmlFor="weight">Weight (kg)</label>
            <input 
              type="number" 
              id="weight"
              className="form-input" 
              step="0.01" 
              min="0"
              value={store.formWeight} 
              onChange={(e) => store.setFormWeight(Number(e.target.value))} 
            />
            {store.fieldErrors.weight && (
              <span className="field-error">{store.fieldErrors.weight}</span>
            )}
          </div>
          <div className="modal-footer">
            <button type="button" className="btn-secondary" onClick={() => store.setIsModalOpen(false)}>Cancel</button>
            <button type="submit" className="btn-primary">Save Product</button>
          </div>
        </form>
      </div>
    </div>
  );
}
