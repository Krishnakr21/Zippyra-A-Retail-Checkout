class Category {
  final String id;
  final String name;
  final String? parentId;
  final int sortOrder;
  final List<Category> children;

  Category({
    required this.id,
    required this.name,
    this.parentId,
    this.sortOrder = 0,
    List<Category>? children,
  }) : children = children ?? [];
}
