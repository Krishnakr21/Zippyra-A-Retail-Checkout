import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/category.dart';

class CategoryChip extends StatelessWidget {
  final Category category;
  final bool isSelected;
  final VoidCallback onTap;

  const CategoryChip({
    super.key,
    required this.category,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8.0),
      child: ChoiceChip(
        label: Text(category.name),
        selected: isSelected,
        onSelected: (_) => onTap(),
        selectedColor: ZippyraColors.primaryBlue.withOpacity(0.2),
        labelStyle: TextStyle(
          color: isSelected ? ZippyraColors.primaryBlue : Colors.black87,
          fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
        ),
      ),
    );
  }
}
