import 'package:path/path.dart';
import 'package:sqflite/sqflite.dart';

class SharedCatalogProduct {
  final String id;
  final String barcode;
  final String name;
  final String description;
  final String? categoryId;
  final int pricePaise;
  final int mrpPaise;
  final String hsnCode;
  final double gstRatePercent;
  final String imageUrl;
  final String thumbnailUrl;
  final bool isReturnable;

  SharedCatalogProduct({
    required this.id,
    required this.barcode,
    required this.name,
    required this.description,
    this.categoryId,
    required this.pricePaise,
    required this.mrpPaise,
    required this.hsnCode,
    required this.gstRatePercent,
    required this.imageUrl,
    required this.thumbnailUrl,
    required this.isReturnable,
  });
}

class SharedCatalogCategory {
  final String id;
  final String name;
  final String? parentId;
  final int sortOrder;
  final List<SharedCatalogCategory> children;

  SharedCatalogCategory({
    required this.id,
    required this.name,
    this.parentId,
    required this.sortOrder,
    List<SharedCatalogCategory>? children,
  }) : children = children ?? [];
}

class CatalogDatabase {
  static final CatalogDatabase instance = CatalogDatabase._internal();
  factory CatalogDatabase() => instance;
  CatalogDatabase._internal();

  Database? _db;

  Future<Database> get database async {
    if (_db != null) return _db!;
    _db = await _initDB();
    return _db!;
  }

  Future<Database> _initDB() async {
    final dbPath = await getDatabasesPath();
    final path = join(dbPath, 'zippyra_catalog.db');

    return await openDatabase(
      path,
      version: 1,
      onCreate: (db, version) async {
        await db.execute('''
          CREATE TABLE catalog_products (
            id TEXT PRIMARY KEY,
            store_id TEXT NOT NULL,
            chain_id TEXT NOT NULL,
            barcode TEXT NOT NULL,
            name TEXT NOT NULL,
            description TEXT,
            category_id TEXT,
            price_paise INTEGER NOT NULL,
            mrp_paise INTEGER NOT NULL,
            hsn_code TEXT NOT NULL,
            gst_rate_percent REAL NOT NULL,
            is_active INTEGER NOT NULL DEFAULT 1,
            is_returnable INTEGER NOT NULL DEFAULT 1,
            image_url TEXT,
            thumbnail_url TEXT,
            sync_seq INTEGER NOT NULL
          );
        ''');

        await db.execute('CREATE UNIQUE INDEX idx_catalog_products_barcode ON catalog_products (store_id, barcode);');
        await db.execute('CREATE INDEX idx_catalog_products_sync ON catalog_products (store_id, sync_seq);');
        await db.execute('CREATE INDEX idx_catalog_products_category ON catalog_products (category_id);');

        await db.execute('''
          CREATE TABLE catalog_categories (
            id TEXT PRIMARY KEY,
            chain_id TEXT NOT NULL,
            name TEXT NOT NULL,
            parent_id TEXT,
            sort_order INTEGER DEFAULT 0
          );
        ''');

        await db.execute('''
          CREATE TABLE catalog_sync_meta (
            store_id TEXT PRIMARY KEY,
            last_sync_seq INTEGER NOT NULL DEFAULT 0,
            last_synced_at TEXT NOT NULL
          );
        ''');
      },
    );
  }

  Future<SharedCatalogProduct?> getProductByBarcode(String storeId, String barcode) async {
    final db = await database;
    final results = await db.query(
      'catalog_products',
      where: 'store_id = ? AND barcode = ? AND is_active = 1',
      whereArgs: [storeId, barcode],
      limit: 1,
    );

    if (results.isEmpty) return null;
    return _mapProduct(results.first);
  }

  Future<List<SharedCatalogProduct>> searchProducts(String storeId, String query, {String? categoryId, int page = 1, int pageSize = 20}) async {
    final db = await database;
    final offset = (page - 1) * pageSize;

    String whereClause = 'store_id = ? AND is_active = 1';
    List<dynamic> whereArgs = [storeId];

    if (query.isNotEmpty) {
      whereClause += ' AND name LIKE ?';
      whereArgs.add('%$query%');
    }

    if (categoryId != null && categoryId.isNotEmpty) {
      whereClause += ' AND category_id = ?';
      whereArgs.add(categoryId);
    }

    final results = await db.query(
      'catalog_products',
      where: whereClause,
      whereArgs: whereArgs,
      orderBy: 'name ASC',
      limit: pageSize,
      offset: offset,
    );

    return results.map(_mapProduct).toList();
  }

  Future<List<SharedCatalogCategory>> getCategories(String chainId) async {
    final db = await database;
    final results = await db.query(
      'catalog_categories',
      where: 'chain_id = ?',
      whereArgs: [chainId],
      orderBy: 'sort_order ASC, name ASC',
    );

    final rawCategories = results.map((row) {
      return SharedCatalogCategory(
        id: row['id'] as String,
        name: row['name'] as String,
        parentId: row['parent_id'] as String?,
        sortOrder: (row['sort_order'] as num?)?.toInt() ?? 0,
      );
    }).toList();

    final Map<String, SharedCatalogCategory> catMap = {};
    for (final c in rawCategories) {
      catMap[c.id] = c;
    }

    final List<SharedCatalogCategory> rootCategories = [];
    for (final c in rawCategories) {
      if (c.parentId == null || c.parentId!.isEmpty) {
        rootCategories.add(c);
      } else if (catMap.containsKey(c.parentId)) {
        catMap[c.parentId]!.children.add(c);
      } else {
        rootCategories.add(c);
      }
    }

    return rootCategories;
  }

  Future<int> getLastSyncSeq(String storeId) async {
    final db = await database;
    final results = await db.query(
      'catalog_sync_meta',
      columns: ['last_sync_seq'],
      where: 'store_id = ?',
      whereArgs: [storeId],
      limit: 1,
    );

    if (results.isEmpty) return 0;
    return (results.first['last_sync_seq'] as num).toInt();
  }

  Future<void> batchWriteSyncPage({
    required String storeId,
    required List<Map<String, dynamic>> productsJson,
    required List<String> deletedIds,
    required int newMaxSeq,
  }) async {
    final db = await database;

    await db.transaction((txn) async {
      final batch = txn.batch();

      for (final deletedId in deletedIds) {
        batch.delete(
          'catalog_products',
          where: 'id = ? AND store_id = ?',
          whereArgs: [deletedId, storeId],
        );
      }

      for (final p in productsJson) {
        batch.insert(
          'catalog_products',
          {
            'id': p['id'],
            'store_id': p['store_id'],
            'chain_id': p['chain_id'],
            'barcode': p['barcode'],
            'name': p['name'],
            'description': p['description'] ?? '',
            'category_id': p['category_id'],
            'price_paise': p['price_paise'],
            'mrp_paise': p['mrp_paise'],
            'hsn_code': p['hsn_code'],
            'gst_rate_percent': p['gst_rate_percent'] ?? 0.0,
            'is_active': (p['is_active'] == true) ? 1 : 0,
            'is_returnable': (p['is_returnable'] == true) ? 1 : 0,
            'image_url': p['image_url'],
            'thumbnail_url': p['thumbnail_url'],
            'sync_seq': p['sync_seq'],
          },
          conflictAlgorithm: ConflictAlgorithm.replace,
        );
      }

      batch.insert(
        'catalog_sync_meta',
        {
          'store_id': storeId,
          'last_sync_seq': newMaxSeq,
          'last_synced_at': DateTime.now().toIso8601String(),
        },
        conflictAlgorithm: ConflictAlgorithm.replace,
      );

      await batch.commit(noResult: true);
    });
  }

  SharedCatalogProduct _mapProduct(Map<String, dynamic> row) {
    return SharedCatalogProduct(
      id: row['id'] as String,
      barcode: row['barcode'] as String,
      name: row['name'] as String,
      description: row['description'] as String? ?? '',
      categoryId: row['category_id'] as String?,
      pricePaise: (row['price_paise'] as num).toInt(),
      mrpPaise: (row['mrp_paise'] as num).toInt(),
      hsnCode: row['hsn_code'] as String? ?? '',
      gstRatePercent: (row['gst_rate_percent'] as num).toDouble(),
      imageUrl: row['image_url'] as String? ?? '',
      thumbnailUrl: row['thumbnail_url'] as String? ?? '',
      isReturnable: (row['is_returnable'] as num?) == 1,
    );
  }
}
