export interface Product {
  id: string;
  name: string;
  sku: string;
  description: string;
  category: string;
  price: number;
  stock: number;
  weight_kg: number;
}

export interface ImportReport {
  total_rows: number;
  imported_rows: number;
  updated_rows: number;
  errors: any[];
}

export interface PurchaseResult {
  id: string;
  customer_id: string;
  sku: string;
  quantity: number;
  total_price: number;
  idempotency_key: string;
  created_at?: string;
}

export const api = {
  async listProducts(query: string = '', category: string = ''): Promise<Product[]> {
    const params = new URLSearchParams();
    if (query) params.append('q', query);
    if (category) params.append('category', category);
    
    const res = await fetch(`/api/products?${params.toString()}`);
    if (!res.ok) throw new Error('Failed to fetch products');
    return res.json();
  },

  async listOrders(): Promise<PurchaseResult[]> {
    const res = await fetch('/api/orders');
    if (!res.ok) throw new Error('Failed to fetch orders');
    return res.json();
  },

  async purchase(sku: string, customer_id: string, quantity: number, expected_price: number, idempotency_key: string): Promise<PurchaseResult> {
    const res = await fetch('/api/purchase', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        sku,
        customer_id,
        quantity,
        expected_price,
        idempotency_key,
      }),
    });
    
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || 'Failed to purchase product');
    }
    return res.json();
  },

  async importCsv(file: File): Promise<ImportReport> {
    const formData = new FormData();
    formData.append('file', file);

    const res = await fetch('/api/products/import', {
      method: 'POST',
      body: formData,
    });

    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || 'Failed to import products');
    }
    return res.json();
  },

  async createProduct(product: Omit<Product, 'id'>): Promise<Product> {
    const res = await fetch('/api/products', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(product),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || 'Failed to create product');
    }
    return res.json();
  },

  async updateProduct(id: string, product: Omit<Product, 'id'>): Promise<Product> {
    const res = await fetch(`/api/products/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(product),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || 'Failed to update product');
    }
    return res.json();
  },

  async deleteProduct(id: string): Promise<void> {
    const res = await fetch(`/api/products/${id}`, {
      method: 'DELETE',
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || 'Failed to delete product');
    }
  }
};
