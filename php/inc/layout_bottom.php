</main>
<footer class="site-footer">
<nav>
<a href="<?= h($legalPrefix) ?>mentions-legales.php">Mentions légales</a>
<a href="<?= h($legalPrefix) ?>confidentialite.php">Politique de confidentialité</a>
<a href="<?= $navMode === 'back' ? '../frontoffice/aide.php' : 'aide.php' ?>">Contact</a>
</nav>
<p>&copy; <?= date('Y') ?> NO MORE WASTE &mdash; Tous droits réservés.</p>
</footer>
</body>
</html>
