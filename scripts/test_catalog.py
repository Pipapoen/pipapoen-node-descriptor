import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('catalog_builder', Path(__file__).with_name('build-catalog.py'))
builder = importlib.util.module_from_spec(spec)
spec.loader.exec_module(builder)


class CatalogTests(unittest.TestCase):
    def test_decimal_and_binary_capacity_boundaries(self):
        self.assertEqual(builder.tier(builder.gib(2.1, 'TB'), builder.DISK), 6)
        self.assertEqual(builder.tier(builder.gib(2.1, 'TiB'), builder.DISK), 7)
        self.assertEqual(builder.gib(4294967296, 'bytes'), 4)
        self.assertEqual(builder.tier(.5, builder.CPU), 0)
        self.assertEqual(builder.tier(257, builder.CPU), 15)

    def test_unknown_or_invalid_capacity_is_rejected(self):
        for value in (0, -1, 'NaN', 'Infinity'):
            with self.assertRaises(ValueError):
                builder.tier(value, builder.RAM)
        with self.assertRaises(KeyError):
            builder.gib(8, 'unknown')

    def test_separately_configured_disks_are_omitted(self):
        plans = builder.generate()['plans.json']
        e2 = next(p for p in plans if p['provider'] == 10 and p['name'] == 'e2-standard-2')
        self.assertNotIn('storage', e2)
        self.assertEqual(e2['cpu'], 1)

    def test_ocpu_conversion_distinguishes_x86_and_arm_a1(self):
        plans = builder.generate()['plans.json']
        x86 = next(p for p in plans if p['name'] == 'VM.Standard.E5.Flex / 2 OCPU / 16 GB')
        arm = next(p for p in plans if p['name'] == 'VM.Standard.A1.Flex / 2 OCPU / 12 GB')
        self.assertEqual((x86['cpu'], x86['architecture']), (3, 0))
        self.assertEqual((arm['cpu'], arm['architecture']), (1, 1))
        self.assertEqual(arm['kind'], 'configuration')


if __name__ == '__main__':
    unittest.main()
