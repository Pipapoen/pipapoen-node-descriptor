import json
import subprocess
import unittest
from unittest.mock import patch

from image_tags import IMAGE, aliases_for, promote


def release(version, **kwargs):
    return dict(tag_name='v' + version, draft=False, prerelease=False, **kwargs)


class ImageTagTests(unittest.TestCase):
    def test_first_stable(self):
        self.assertEqual(aliases_for('26.1.0', [release('26.1.0')]), ['26.1', '26', 'latest'])

    def test_patch_forward(self):
        self.assertEqual(aliases_for('26.2.3', [release('26.2.2'), release('26.2.3')]), ['26.2', '26', 'latest'])

    def test_old_series_patch(self):
        self.assertEqual(aliases_for('26.2.4', [release('26.3.0'), release('26.2.4')]), ['26.2'])

    def test_old_year_patch(self):
        self.assertEqual(aliases_for('26.2.4', [release('27.1.0'), release('26.2.4')]), ['26.2', '26'])

    def test_old_patch_cannot_move_any_alias(self):
        self.assertEqual(aliases_for('26.2.3', [release('26.2.4'), release('26.2.3')]), [])

    def test_numeric_order(self):
        self.assertEqual(aliases_for('26.10.0', [release('26.9.99'), release('26.10.0')]), ['26.10', '26', 'latest'])

    def test_unpublished_and_candidate_releases_do_not_advance_aliases(self):
        rows = [release('26.1.0'), dict(tag_name='v27.1.0', draft=True, prerelease=False),
                dict(tag_name='v28.1.0-rc.1', draft=False, prerelease=True)]
        self.assertEqual(aliases_for('26.1.0', rows), ['26.1', '26', 'latest'])

    def test_missing_current_release_fails(self):
        with self.assertRaises(ValueError):
            aliases_for('26.2.0', [release('26.1.0')])

    @patch('image_tags.subprocess.run')
    def test_candidate_never_promotes(self, run):
        self.assertEqual(promote('26.3.0-rc.1', 'unused'), [])
        run.assert_not_called()

    @patch('image_tags.subprocess.run')
    def test_promote_uses_all_pages_and_published_digest(self, run):
        run.return_value.stdout = json.dumps([[release('26.2.4')], [release('27.1.0')]])
        digest = 'sha256:' + 'a' * 64
        self.assertEqual(promote('26.2.4', digest), ['26.2', '26'])
        self.assertEqual(run.call_args_list[1].args[0], [
            'docker', 'buildx', 'imagetools', 'create', '--prefer-index=false',
            '--tag', IMAGE + ':26.2', '--tag', IMAGE + ':26', IMAGE + '@' + digest,
        ])
        self.assertEqual(run.call_count, 2)

    @patch('image_tags.subprocess.run')
    def test_failed_release_query_does_not_write(self, run):
        run.side_effect = subprocess.CalledProcessError(1, 'gh')
        with self.assertRaises(subprocess.CalledProcessError):
            promote('26.1.0', 'sha256:' + 'a' * 64)
        self.assertEqual(run.call_count, 1)

    @patch('image_tags.subprocess.run')
    def test_latest_release_tracks_successful_image_promotion(self, run):
        run.return_value.stdout = json.dumps([[release('26.2.0')]])
        promote('26.2.0', 'sha256:' + 'a' * 64)
        self.assertIn('--latest', run.call_args_list[-1].args[0])


if __name__ == '__main__':
    unittest.main()
