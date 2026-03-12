import random, math
import numpy as np
import matplotlib.pyplot as plt
import matplotlib.colors as pltc
from scipy.stats import qmc
from scipy.spatial import Voronoi, voronoi_plot_2d


def norm1_distance(a, b):
    return abs(a[0] - b[0]) + abs(a[1] - b[1])


def norm2_distance(a, b):
    return math.sqrt((a[0] - b[0]) ** 2 + (a[1] - b[1]) ** 2)


class WorldGenerator:
    world_w = 128
    world_h = 128
    i_min = 10
    isl_size = 7
    m_cities = 3
    max_mountain = 5
    p_cities = 6
    max_plains = 3
    c_cities = 6
    max_coast = 3

    def __init__(self):
        self.world = [[0 for _ in range(self.world_w)] for _ in range(self.world_h)]

    def generate_map(self):
        self.centers = self.generate_island_centers()
        for center in self.centers:
            x, y = center
            self.world[y][x] = 5

    def generate_island_centers(self, k=30) -> list[tuple]:
        """
        Fast Poisson Disk Sampling in Arbitrary Dimensions

        Args:
            k (int, optional): Samples before rejection. Defaults to 30.
        Returns:
            list[tuple]: Valid center list.
        """

        rng = np.random.default_rng()
        engine = qmc.PoissonDisk(
            d=2, radius=self.i_min / max(self.world_w, self.world_h), 
            rng=rng, ncandidates=k
        )
        sample = engine.fill_space()

        centers = []
        for x, y in sample:
            cx = int(x * self.world_w)
            cy = int(y * self.world_h)
            centers.append((cx, cy))

        return centers

    def generate_island(self):
        pass

    def show_map(self):
        cvals = [1, 2, 3, 4, 5]
        colors = [
            "royalblue",
            "cornflowerblue",
            "navajowhite",
            "forestgreen",
            "black"# "dimgrey",
        ]

        norm = plt.Normalize(min(cvals), max(cvals))
        tuples = list(zip(map(norm, cvals), colors))
        cmap = pltc.LinearSegmentedColormap.from_list("", tuples)
        cmap = pltc.ListedColormap(colors)

        fig, ax = plt.subplots()

        # plot island centers with minimum distance
        for xi, yi in self.centers:
            circle = plt.Circle((xi, yi), radius=self.i_min/2-1, fill=False)    
            ax.add_artist(circle)

        # plot world map
        ax.imshow(self.world, cmap=cmap, vmin=0.5, vmax=5.5)
        ax.set_xlim(0, self.world_w-1)
        ax.set_ylim(0, self.world_h-1)

        # plot voronoi regions
        vor = Voronoi(self.centers)
        voronoi_plot_2d(vor, ax=ax, show_vertices=False, line_colors='black')

        # plt.colorbar()
        plt.title("World Map")
        plt.show()

    def run(self):
        self.generate_map()
        self.show_map()


def run():
    generator = WorldGenerator()
    generator.run()


if __name__ == "__main__":
    run()
