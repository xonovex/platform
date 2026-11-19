import unittest

from src.cap import apply_reduction, cap_daily_total
from src.export import export_capped_totals


class CapTest(unittest.TestCase):
    def test_sums_the_fares_below_the_daily_price(self):
        self.assertEqual(cap_daily_total([250, 300], 1000), 550)

    def test_caps_the_total_at_the_daily_price(self):
        self.assertEqual(cap_daily_total([600, 700], 1000), 1000)

    def test_rounds_the_reduced_total_down_to_whole_cents(self):
        self.assertEqual(apply_reduction(1001, 15), 850)

    def test_exports_a_header_and_one_line_for_each_day(self):
        days = [{"date": "2026-10-01", "fares": [600, 700]}]
        self.assertEqual(export_capped_totals(days, 1000), "date,total\n2026-10-01,1000\n")


if __name__ == "__main__":
    unittest.main()
